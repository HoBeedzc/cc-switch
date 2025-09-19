package export

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"time"

	"cc-switch/internal/common"
)

const (
	// CCX file format magic number
	MagicNumber = "CCX1"
	Version     = 1

	// Flags
	FlagEncrypted  = 1 << 0
	FlagCompressed = 1 << 1
)

// CCXHeader represents the CCX file header
type CCXHeader struct {
	Magic     [4]byte // "CCX1"
	Version   uint32  // Format version
	Flags     uint32  // Feature flags
	Timestamp int64   // Unix timestamp
	DataLen   uint64  // Length of data section
	Checksum  uint32  // CRC32 checksum
}

// CCXMetadata contains file metadata
type CCXMetadata struct {
	Version       string `json:"version"`
	ExportedAt    string `json:"exported_at"`
	ToolVersion   string `json:"tool_version"`
	ExportType    string `json:"export_type"`
	ProfilesCount int    `json:"profiles_count"`
	Encryption    string `json:"encryption"`
	Compression   string `json:"compression"`
}

// ProfileData represents a profile in the export
type ProfileData struct {
	Name      string                 `json:"name"`
	IsCurrent bool                   `json:"is_current"`
	Content   map[string]interface{} `json:"content"`
	Metadata  ProfileMetadata        `json:"metadata"`
}

// ProfileMetadata contains profile metadata
type ProfileMetadata struct {
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at"`
}

// ExportData represents the complete export structure
type ExportData struct {
	Profiles []ProfileData `json:"profiles"`
}

// CCXHandler handles CCX file format operations
type CCXHandler struct{}

// NewCCXHandler creates a new CCX format handler
func NewCCXHandler() *CCXHandler {
	return &CCXHandler{}
}

// Write writes export data to CCX format
func (h *CCXHandler) Write(data *ExportData, writer io.Writer, password string) error {
	// Create metadata
	metadata := CCXMetadata{
		Version:       "1.0.0",
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		ToolVersion:   "cc-switch v1.0.0",
		ExportType:    h.getExportType(data),
		ProfilesCount: len(data.Profiles),
		Encryption:    "aes-256-gcm",
		Compression:   "gzip",
	}

	// Serialize metadata
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	// Serialize payload
	payloadBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	// Compress payload
	compressedPayload, err := common.CompressData(payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to compress payload: %w", err)
	}

	// Encrypt payload if password provided
	var finalPayload []byte
	flags := uint32(FlagCompressed)

	if password != "" {
		encData, err := common.EncryptData(compressedPayload, password)
		if err != nil {
			return fmt.Errorf("failed to encrypt payload: %w", err)
		}

		// Serialize encryption data
		encBytes, err := h.serializeEncryptionData(encData)
		if err != nil {
			return fmt.Errorf("failed to serialize encryption data: %w", err)
		}

		finalPayload = encBytes
		flags |= FlagEncrypted
	} else {
		finalPayload = compressedPayload
	}

	// Prepare metadata with length prefix for checksum calculation
	var metadataWithLength bytes.Buffer
	if err := h.writeWithLength(&metadataWithLength, metadataBytes); err != nil {
		return fmt.Errorf("failed to prepare metadata for checksum: %w", err)
	}
	metadataWithLengthBytes := metadataWithLength.Bytes()

	// Calculate checksum
	checksum := crc32.ChecksumIEEE(append(metadataWithLengthBytes, finalPayload...))

	// Create header
	header := CCXHeader{
		Version:   Version,
		Flags:     flags,
		Timestamp: time.Now().Unix(),
		DataLen:   uint64(len(metadataWithLengthBytes) + len(finalPayload)),
		Checksum:  checksum,
	}
	copy(header.Magic[:], MagicNumber)

	// Write header
	if err := binary.Write(writer, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write metadata with length prefix
	if _, err := writer.Write(metadataWithLengthBytes); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Write payload
	if _, err := writer.Write(finalPayload); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	return nil
}

// Read reads export data from CCX format
func (h *CCXHandler) Read(reader io.Reader, password string) (*ExportData, error) {
	// Read and validate header
	var header CCXHeader
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if string(header.Magic[:]) != MagicNumber {
		return nil, fmt.Errorf("invalid file format: magic number mismatch")
	}

	if header.Version != Version {
		return nil, fmt.Errorf("unsupported file version: %d", header.Version)
	}

	// Read metadata
	metadataBytes, err := h.readWithLength(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Read payload
	payloadSize := header.DataLen - uint64(len(metadataBytes)) - 4 // 4 bytes for metadata length
	payloadBytes := make([]byte, payloadSize)
	if _, err := io.ReadFull(reader, payloadBytes); err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	// Verify checksum
	// Prepare metadata with length prefix for checksum verification
	var metadataWithLength bytes.Buffer
	if err := h.writeWithLength(&metadataWithLength, metadataBytes); err != nil {
		return nil, fmt.Errorf("failed to prepare metadata for checksum verification: %w", err)
	}
	metadataWithLengthBytes := metadataWithLength.Bytes()

	expectedChecksum := crc32.ChecksumIEEE(append(metadataWithLengthBytes, payloadBytes...))
	if expectedChecksum != header.Checksum {
		return nil, fmt.Errorf("file integrity check failed: checksum mismatch")
	}

	// Decrypt payload if encrypted
	var compressedPayload []byte
	if header.Flags&FlagEncrypted != 0 {
		if password == "" {
			return nil, fmt.Errorf("file is encrypted but no password provided")
		}

		encData, err := h.deserializeEncryptionData(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize encryption data: %w", err)
		}

		decrypted, err := common.DecryptData(encData, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt payload: %w", err)
		}

		compressedPayload = decrypted
	} else {
		compressedPayload = payloadBytes
	}

	// Decompress payload if compressed
	var finalPayload []byte
	if header.Flags&FlagCompressed != 0 {
		decompressed, err := common.DecompressData(compressedPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress payload: %w", err)
		}
		finalPayload = decompressed
	} else {
		finalPayload = compressedPayload
	}

	// Parse export data
	var exportData ExportData
	if err := json.Unmarshal(finalPayload, &exportData); err != nil {
		return nil, fmt.Errorf("failed to parse export data: %w", err)
	}

	return &exportData, nil
}

// ValidateFile validates CCX file format without reading the payload
func (h *CCXHandler) ValidateFile(reader io.Reader) (*CCXMetadata, error) {
	// Read and validate header
	var header CCXHeader
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if string(header.Magic[:]) != MagicNumber {
		return nil, fmt.Errorf("invalid file format: magic number mismatch")
	}

	if header.Version != Version {
		return nil, fmt.Errorf("unsupported file version: %d", header.Version)
	}

	// Read metadata
	metadataBytes, err := h.readWithLength(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Parse metadata
	var metadata CCXMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}

// Helper methods

func (h *CCXHandler) getExportType(data *ExportData) string {
	if len(data.Profiles) == 1 {
		return "single"
	}
	return "multiple"
}

func (h *CCXHandler) writeWithLength(writer io.Writer, data []byte) error {
	// Write length prefix (4 bytes)
	length := uint32(len(data))
	if err := binary.Write(writer, binary.LittleEndian, length); err != nil {
		return err
	}

	// Write data
	_, err := writer.Write(data)
	return err
}

func (h *CCXHandler) readWithLength(reader io.Reader) ([]byte, error) {
	// Read length prefix
	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, err
	}

	return data, nil
}

func (h *CCXHandler) serializeEncryptionData(encData *common.EncryptionData) ([]byte, error) {
	var buf bytes.Buffer

	// Write salt length and salt
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(encData.Salt))); err != nil {
		return nil, err
	}
	buf.Write(encData.Salt)

	// Write nonce length and nonce
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(encData.Nonce))); err != nil {
		return nil, err
	}
	buf.Write(encData.Nonce)

	// Write encrypted data length and data
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(encData.Encrypted))); err != nil {
		return nil, err
	}
	buf.Write(encData.Encrypted)

	return buf.Bytes(), nil
}

func (h *CCXHandler) deserializeEncryptionData(data []byte) (*common.EncryptionData, error) {
	reader := bytes.NewReader(data)

	// Read salt
	var saltLen uint32
	if err := binary.Read(reader, binary.LittleEndian, &saltLen); err != nil {
		return nil, err
	}
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return nil, err
	}

	// Read nonce
	var nonceLen uint32
	if err := binary.Read(reader, binary.LittleEndian, &nonceLen); err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(reader, nonce); err != nil {
		return nil, err
	}

	// Read encrypted data
	var encLen uint32
	if err := binary.Read(reader, binary.LittleEndian, &encLen); err != nil {
		return nil, err
	}
	encrypted := make([]byte, encLen)
	if _, err := io.ReadFull(reader, encrypted); err != nil {
		return nil, err
	}

	return &common.EncryptionData{
		Salt:      salt,
		Nonce:     nonce,
		Encrypted: encrypted,
	}, nil
}
