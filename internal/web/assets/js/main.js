class CCSwitch {
    constructor() {
        this.currentProfile = null;
        this.profiles = [];
        this.templates = [];
        this.isEmptyMode = false;
        this.toastContainer = null;
        this.toasts = new Map();
        this.toastId = 0;
        this.init();
    }

    async init() {
        console.log('🚀 Initializing cc-switch web interface...');
        
        // Setup navigation
        this.setupNavigation();
        
        // Load initial data
        await this.loadData();
        
        // Setup event listeners
        this.setupEventListeners();
        
        // Check for updates
        this.checkForUpdates();
        
        // Show profiles tab by default
        this.showSection('profiles');
        
        console.log('✅ cc-switch web interface ready!');
    }

    setupNavigation() {
        const navTabs = document.querySelectorAll('.nav-tab');
        
        navTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                const section = e.target.dataset.section;
                this.showSection(section);
            });
        });
    }

    showSection(sectionName) {
        // Update nav tabs
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.section === sectionName);
        });

        // Update sections
        document.querySelectorAll('.section').forEach(section => {
            section.classList.toggle('active', section.id === `${sectionName}-section`);
        });

        // Load section-specific data
        switch (sectionName) {
            case 'profiles':
                this.renderProfiles();
                break;
            case 'templates':
                this.renderTemplates();
                break;
            case 'settings':
                this.renderSettings();
                break;
            case 'test':
                this.renderTest();
                break;
        }
    }

    async loadData() {
        try {
            // Load profiles
            await this.loadProfiles();
            
            // Load templates
            await this.loadTemplates();
            
            // Load current configuration
            await this.loadCurrentConfig();
            
            // Load health status
            await this.loadHealthStatus();
            
        } catch (error) {
            console.error('Failed to load data:', error);
            this.showError('Failed to load configuration data');
        }
    }

    async loadProfiles() {
        try {
            const response = await this.apiCall('/api/profiles');
            this.profiles = response.data.profiles || [];
        } catch (error) {
            console.error('Failed to load profiles:', error);
            this.profiles = [];
        }
    }

    async loadCurrentConfig() {
        try {
            const response = await this.apiCall('/api/current');
            this.currentProfile = response.data.current;
            this.isEmptyMode = response.data.empty_mode;
        } catch (error) {
            console.error('Failed to load current config:', error);
        }
    }

    async loadHealthStatus() {
        try {
            const response = await this.apiCall('/api/health');
            console.log('Health status:', response.data);
        } catch (error) {
            console.error('Failed to load health status:', error);
        }
    }

    async loadTemplates() {
        try {
            const response = await this.apiCall('/api/templates');
            this.templates = response.data.templates || [];
            console.log('Loaded templates:', this.templates);
        } catch (error) {
            console.error('Failed to load templates:', error);
            this.templates = [];
        }
    }

    renderProfiles() {
        const container = document.getElementById('profiles-list');
        if (!container) return;

        if (this.profiles.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <h3>No configurations found</h3>
                    <p>Create your first configuration to get started.</p>
                    <button class="btn btn-primary mt-4" onclick="app.createProfile()">
                        Create Configuration
                    </button>
                </div>
            `;
            return;
        }

        const profilesHTML = this.profiles.map(profile => {
            const isCurrent = profile.name === this.currentProfile && !this.isEmptyMode;
            
            return `
                <div class="profile-item ${isCurrent ? 'current' : ''}">
                    <div class="profile-info">
                        <div class="profile-name">${this.escapeHtml(profile.name)}</div>
                        ${isCurrent ? '<div class="profile-status current">Current</div>' : ''}
                    </div>
                    <div class="profile-actions">
                        ${!isCurrent ? `<button class="btn btn-success" onclick="app.switchProfile('${this.escapeHtml(profile.name)}')">Use</button>` : ''}
                        <button class="btn btn-outline" onclick="app.viewProfile('${this.escapeHtml(profile.name)}')">View</button>
                        <button class="btn btn-warning" onclick="app.editProfile('${this.escapeHtml(profile.name)}')">Edit</button>
                        <button class="btn btn-danger" onclick="app.deleteProfile('${this.escapeHtml(profile.name)}')">Delete</button>
                    </div>
                </div>
            `;
        }).join('');

        container.innerHTML = `
            <div class="flex justify-between items-center mb-4">
                <h2>Available Configurations</h2>
                <div class="flex gap-2">
                    ${this.isEmptyMode ? '<button class="btn btn-success" onclick="app.restoreFromEmptyMode()">Restore Config</button>' : '<button class="btn btn-secondary" onclick="app.useEmptyMode()">Empty Mode</button>'}
                    <button class="btn btn-primary" onclick="app.createProfile()">New Config</button>
                </div>
            </div>
            ${this.isEmptyMode ? '<div class="status status-offline">⚠️ Empty mode active (no configuration active)</div>' : ''}
            <div class="profile-list">
                ${profilesHTML}
            </div>
        `;
    }

    renderSettings() {
        const container = document.getElementById('settings-content');
        if (!container) return;

        container.innerHTML = `
            <h2>Settings & Configuration</h2>
            <div class="form-group">
                <label class="form-label">Current Profile</label>
                <p>${this.isEmptyMode ? 'None (Empty Mode)' : this.currentProfile || 'None'}</p>
            </div>
            <div class="form-group">
                <label class="form-label">Configuration Directory</label>
                <p><code>~/.claude/profiles/</code></p>
            </div>
            <div class="form-group">
                <button class="btn btn-primary" onclick="app.exportConfigs()">Export All Configs</button>
                <button class="btn btn-outline" onclick="app.importConfigs()">Import Configs</button>
            </div>
        `;
    }

    renderTest() {
        const container = document.getElementById('test-content');
        if (!container) return;

        container.innerHTML = `
            <h2>API Connectivity Testing</h2>
            <div class="form-group">
                <label class="form-label">Profile to Test</label>
                <select id="test-profile" class="form-input">
                    <option value="">Current Configuration</option>
                    ${this.profiles.map(p => `<option value="${this.escapeHtml(p.name)}">${this.escapeHtml(p.name)}</option>`).join('')}
                </select>
            </div>
            <div class="form-group">
                <label>
                    <input type="checkbox" id="test-quick"> Quick Test
                </label>
            </div>
            <div class="form-group">
                <button class="btn btn-primary" onclick="app.runConnectivityTest()" id="test-button">
                    Run Test
                </button>
            </div>
            <div id="test-results" style="display: none;">
                <h3>Test Results</h3>
                <div id="test-results-content"></div>
            </div>
        `;
    }

    renderTemplates() {
        const container = document.getElementById('templates-list');
        if (!container) return;

        if (this.templates.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <h3>No templates found</h3>
                    <p>Create your first template to get started.</p>
                    <button class="btn btn-primary mt-4" onclick="app.createTemplate()">
                        Create Template
                    </button>
                </div>
            `;
            return;
        }

        const templatesHTML = this.templates.map(template => {
            const isDefault = template === 'default';
            
            return `
                <div class="profile-item template-item">
                    <div class="profile-info">
                        <div class="profile-name">${this.escapeHtml(template)}</div>
                        ${isDefault ? '<div class="profile-status system">System Default</div>' : ''}
                    </div>
                    <div class="profile-actions">
                        <button class="btn btn-outline" onclick="app.viewTemplate('${this.escapeHtml(template)}')">View</button>
                        ${!isDefault ? `<button class="btn btn-warning" onclick="app.editTemplate('${this.escapeHtml(template)}')">Edit</button>` : ''}
                        ${!isDefault ? `<button class="btn btn-secondary" onclick="app.copyTemplate('${this.escapeHtml(template)}')">Copy</button>` : ''}
                        ${!isDefault ? `<button class="btn btn-danger" onclick="app.deleteTemplate('${this.escapeHtml(template)}')">Delete</button>` : ''}
                        <button class="btn btn-primary" onclick="app.createProfileFromTemplate('${this.escapeHtml(template)}')">Create Config</button>
                    </div>
                </div>
            `;
        }).join('');

        container.innerHTML = `
            <div class="flex justify-between items-center mb-4">
                <h2>Available Templates</h2>
                <div class="flex gap-2">
                    <button class="btn btn-secondary" onclick="app.loadTemplates(); app.renderTemplates();">Refresh</button>
                    <button class="btn btn-primary" onclick="app.createTemplate()">+ Create Template</button>
                </div>
            </div>
            <div class="profile-list">
                ${templatesHTML}
            </div>
        `;
    }

    // API Methods
    async switchProfile(profileName) {
        try {
            const response = await this.apiCall('/api/switch', {
                method: 'POST',
                body: JSON.stringify({ profile: profileName })
            });
            
            this.showSuccess(`Switched to configuration: ${profileName}`);
            await this.loadData();
            this.renderProfiles();
        } catch (error) {
            this.showError(`Failed to switch profile: ${error.message}`);
        }
    }

    async useEmptyMode() {
        try {
            const response = await this.apiCall('/api/switch', {
                method: 'POST',
                body: JSON.stringify({ profile: '' })
            });
            
            this.showSuccess('Switched to empty mode');
            await this.loadData();
            this.renderProfiles();
        } catch (error) {
            this.showError(`Failed to switch to empty mode: ${error.message}`);
        }
    }

    async restoreFromEmptyMode() {
        try {
            const response = await this.apiCall('/api/switch', {
                method: 'POST',
                body: JSON.stringify({ restore: true })
            });
            
            this.showSuccess('Configuration restored from empty mode');
            await this.loadData();
            this.renderProfiles();
        } catch (error) {
            this.showError(`Failed to restore from empty mode: ${error.message}`);
        }
    }

    async deleteProfile(profileName) {
        const confirmed = await this.showConfirm(
            `Are you sure you want to delete configuration "${profileName}"?`,
            {
                title: 'Delete Configuration',
                type: 'danger',
                confirmText: 'Delete',
                confirmClass: 'btn-danger'
            }
        );
        
        if (!confirmed) return;

        try {
            const response = await this.apiCall(`/api/profiles/${encodeURIComponent(profileName)}`, {
                method: 'DELETE'
            });
            
            this.showSuccess(`Configuration "${profileName}" deleted successfully`);
            await this.loadData();
            this.renderProfiles();
        } catch (error) {
            this.showError(`Failed to delete profile: ${error.message}`);
        }
    }

    async runConnectivityTest() {
        const profileSelect = document.getElementById('test-profile');
        const quickCheck = document.getElementById('test-quick');
        const testButton = document.getElementById('test-button');
        const resultsDiv = document.getElementById('test-results');
        const resultsContent = document.getElementById('test-results-content');

        const profile = profileSelect.value;
        const quick = quickCheck.checked;

        testButton.disabled = true;
        testButton.innerHTML = '<div class="spinner"></div>Testing...';

        try {
            const response = await this.apiCall('/api/test', {
                method: 'POST',
                body: JSON.stringify({
                    profile: profile,
                    quick: quick,
                    timeout: 45
                })
            });

            const result = response.data;

            // Build detailed test results display
            let testDetailsHTML = '';
            if (result.tests && result.tests.length > 0) {
                testDetailsHTML = result.tests.map(test => {
                    const statusIcon = test.status === 'success' ? '✅' :
                                     test.status === 'timeout' ? '⏱️' : '❌';
                    const statusClass = test.status === 'success' ? 'status-online' : 'status-offline';
                    const responseTime = test.response_time_ms ?
                        Math.round(test.response_time_ms / 1000000) : 0;

                    return `
                        <div class="test-detail" style="margin-bottom: 1rem; padding: 1rem; border-left: 3px solid ${test.status === 'success' ? '#28a745' : '#dc3545'}; background: var(--bg-secondary);">
                            <div class="${statusClass}" style="margin-bottom: 0.5rem;">
                                ${statusIcon} <strong>${this.getTestName(test)}</strong> (${responseTime}ms)
                            </div>
                            <div style="font-size: 0.9rem; color: var(--text-secondary);">
                                <div><strong>Endpoint:</strong> ${test.endpoint}</div>
                                <div><strong>Method:</strong> ${test.method}</div>
                                ${test.status_code ? `<div><strong>Status Code:</strong> ${test.status_code}</div>` : ''}
                                <div><strong>Response Time:</strong> ${responseTime}ms</div>
                                ${test.details ? `<div><strong>Details:</strong> ${test.details}</div>` : ''}
                                ${test.error ? `<div style="color: #dc3545;"><strong>Error:</strong> ${test.error}</div>` : ''}
                            </div>
                        </div>
                    `;
                }).join('');
            }

            resultsContent.innerHTML = `
                <div class="status ${result.is_connectable ? 'status-online' : 'status-offline'}" style="margin-bottom: 1rem;">
                    ${result.is_connectable ? '✅ CONNECTED' : '❌ CONNECTION FAILED'}
                </div>
                <div style="margin-bottom: 1rem;">
                    <p><strong>Profile:</strong> ${result.profile_name}</p>
                    <p><strong>Response Time:</strong> ${Math.round(result.response_time_ms / 1000000)}ms</p>
                    <p><strong>Tested At:</strong> ${new Date(result.tested_at).toLocaleString()}</p>
                </div>
                ${result.error ? `<div class="status-offline" style="margin-bottom: 1rem;"><strong>Error:</strong> ${result.error}</div>` : ''}
                ${testDetailsHTML ? `
                    <div class="test-details">
                        <h4 style="margin-bottom: 1rem;">Test Details:</h4>
                        ${testDetailsHTML}
                    </div>
                ` : ''}
                <div class="status ${result.is_connectable ? 'status-online' : 'status-offline'}" style="margin-top: 1rem;">
                    ${result.is_connectable ? '✅ Result: Configuration is functional' : '❌ Result: Configuration has issues'}
                    <br><small>Total response time: ${Math.round(result.response_time_ms / 1000000 / 1000)}s</small>
                </div>
            `;
            
            resultsDiv.style.display = 'block';
        } catch (error) {
            this.showError(`Test failed: ${error.message}`);
        } finally {
            testButton.disabled = false;
            testButton.innerHTML = 'Run Test';
        }
    }

    // Helper function to get user-friendly test names
    getTestName(test) {
        if (test.method === 'GET' && test.endpoint === '/v1/models') {
            return 'Authentication Test';
        } else if (test.method === 'GET-MODELS' && test.endpoint === '/v1/models') {
            return 'Models Endpoint';
        } else if (test.method === 'claude-cli' && test.endpoint === '/v1/messages') {
            return 'Chat Endpoint (Claude CLI)';
        } else if (test.method === 'HEAD') {
            return 'Basic Connectivity';
        } else {
            return `${test.method} ${test.endpoint}`;
        }
    }

    // Placeholder methods for implementation
    async viewProfile(profileName) {
        try {
            const response = await this.apiCall(`/api/profiles/${encodeURIComponent(profileName)}`);
            this.showViewModal(response.data);
        } catch (error) {
            this.showError(`Failed to load profile: ${error.message}`);
        }
    }

    async editProfile(profileName) {
        try {
            const response = await this.apiCall(`/api/profiles/${encodeURIComponent(profileName)}`);
            this.showEditModal(response.data);
        } catch (error) {
            this.showError(`Failed to load profile for editing: ${error.message}`);
        }
    }

    showViewModal(profile) {
        const modal = this.createModal(`View Profile: ${profile.name}`, this.renderProfileView(profile));
        document.body.appendChild(modal);
    }

    showEditModal(profile) {
        const modal = this.createModal(`Edit Profile: ${profile.name}`, this.renderProfileEdit(profile), [
            { text: 'Cancel', class: 'btn-secondary', onclick: () => this.closeModal() },
            { text: 'Save Changes', class: 'btn-primary', onclick: () => this.saveProfileChanges(profile.name) }
        ]);
        document.body.appendChild(modal);
        
        // Initialize edit mode state
        window.currentEditMode = 'form'; // 'form' or 'raw'
        window.currentProfileData = profile;
    }

    createModal(title, content, buttons = null) {
        const overlay = document.createElement('div');
        overlay.className = 'modal-overlay';
        
        const modal = document.createElement('div');
        modal.className = 'modal-content';
        
        const header = document.createElement('div');
        header.className = 'modal-header';
        header.innerHTML = `
            <h2>${this.escapeHtml(title)}</h2>
            <button class="modal-close" onclick="app.closeModal()">&times;</button>
        `;
        
        const body = document.createElement('div');
        body.className = 'modal-body';
        body.innerHTML = content;
        
        modal.appendChild(header);
        modal.appendChild(body);
        
        if (buttons) {
            const footer = document.createElement('div');
            footer.className = 'modal-footer';
            buttons.forEach(btn => {
                const button = document.createElement('button');
                button.className = `btn ${btn.class}`;
                button.textContent = btn.text;
                button.onclick = btn.onclick;
                if (btn.id) {
                    button.id = btn.id;
                }
                if (btn.disabled) {
                    button.disabled = btn.disabled;
                }
                footer.appendChild(button);
            });
            modal.appendChild(footer);
        }
        
        overlay.appendChild(modal);
        
        // Close on overlay click
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) {
                this.closeModal();
            }
        });
        
        // Close on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeModal();
            }
        });
        
        return overlay;
    }

    closeModal() {
        const modal = document.querySelector('.modal-overlay');
        if (modal) {
            modal.remove();
        }
    }

    // Simple modal for view-only content
    showModal(title, content, buttonText = 'Close', onSave = null) {
        const buttons = onSave ? [
            { text: 'Cancel', class: 'btn-secondary', onclick: () => this.closeModal() },
            { text: buttonText, class: 'btn-primary', onclick: async () => {
                const result = await onSave();
                if (result !== false) {
                    this.closeModal();
                }
            }}
        ] : [
            { text: buttonText, class: 'btn-primary', onclick: () => this.closeModal() }
        ];
        
        const modal = this.createModal(title, content, buttons);
        document.body.appendChild(modal);
    }

    renderProfileView(profile) {
        const content = profile.content;
        const formattedJson = JSON.stringify(content, null, 2);
        
        return `
            <div class="profile-metadata">
                <dt>Name:</dt>
                <dd>${this.escapeHtml(profile.name)}</dd>
                <dt>Path:</dt>
                <dd>${this.escapeHtml(profile.path)}</dd>
                <dt>Current:</dt>
                <dd>${profile.is_current ? 'Yes' : 'No'}</dd>
            </div>
            
            <h3>Configuration Content:</h3>
            <div class="code-block">${this.syntaxHighlight(formattedJson)}</div>
        `;
    }

    renderProfileEdit(profile) {
        const content = profile.content;
        
        return `
            <div class="form-group">
                <label class="form-label">Profile Name</label>
                <input type="text" id="profile-name-input" value="${this.escapeHtml(profile.name)}" class="form-input">
                <small style="color: var(--text-secondary); display: block; margin-top: 0.25rem;">
                    Enter new name to rename profile
                </small>
            </div>
            
            <div class="profile-metadata">
                <dt>Path:</dt>
                <dd>${this.escapeHtml(profile.path)}</dd>
            </div>
            
            <!-- Edit Mode Toggle -->
            <div class="edit-mode-toggle" style="margin-bottom: 1rem;">
                <div class="flex items-center gap-6">
                    <label class="edit-mode-label" style="display: flex; align-items: center; cursor: pointer;">
                        <input type="radio" name="edit-mode" value="form" checked onchange="app.toggleEditMode('form')" style="margin-right: 0.5rem;">
                        Form Mode
                    </label>
                    <label class="edit-mode-label" style="display: flex; align-items: center; cursor: pointer;">
                        <input type="radio" name="edit-mode" value="raw" onchange="app.toggleEditMode('raw')" style="margin-right: 0.5rem;">
                        Raw JSON
                    </label>
                </div>
            </div>
            
            <!-- Form Editor -->
            <div id="form-editor" style="display: block;">
                <form id="profile-edit-form">
                    ${this.renderEditSection('Environment Variables', 'env', content.env || {})}
                    ${this.renderEditSection('Permissions - Allow', 'permissions_allow', content.permissions?.allow || [])}
                    ${this.renderEditSection('Permissions - Deny', 'permissions_deny', content.permissions?.deny || [])}
                    ${this.renderEditSection('Status Line', 'statusLine', content.statusLine || {})}
                </form>
            </div>
            
            <!-- Raw JSON Editor -->
            <div id="raw-editor" style="display: none;">
                <div class="form-group">
                    <label class="form-label">JSON Configuration</label>
                    <div style="position: relative;">
                        <div id="raw-json-display" class="code-block" style="margin: 0; min-height: 400px; position: relative; z-index: 1;">
                            ${this.syntaxHighlight(JSON.stringify(content, null, 2))}
                        </div>
                        <textarea id="raw-json-textarea" style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; min-height: 400px; font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace; font-size: 0.875rem; line-height: 1.6; padding: 1rem; background: transparent; border: none; outline: none; resize: vertical; color: transparent; caret-color: #e5e7eb; z-index: 2;">${JSON.stringify(content, null, 2)}</textarea>
                    </div>
                    <small style="color: var(--text-secondary); display: block; margin-top: 0.25rem;">
                        Edit the configuration as raw JSON. Make sure the syntax is valid.
                    </small>
                </div>
                <div class="json-validation" id="json-validation-message" style="display: none; margin-top: 0.5rem;"></div>
            </div>
        `;
    }

    renderEditSection(title, key, data) {
        if (key === 'permissions_allow' || key === 'permissions_deny') {
            return this.renderArrayEditor(title, key, data);
        } else if (typeof data === 'object') {
            return this.renderObjectEditor(title, key, data);
        }
        return '';
    }

    renderObjectEditor(title, key, obj) {
        const entries = Object.entries(obj);
        
        return `
            <div class="form-section">
                <div class="form-section-title">${title}</div>
                <div class="kv-editor" data-section="${key}">
                    ${entries.map(([k, v], index) => `
                        <div class="kv-item">
                            <input type="text" value="${this.escapeHtml(k)}" placeholder="Key" data-type="key">
                            <input type="text" value="${this.escapeHtml(String(v))}" placeholder="Value" data-type="value">
                            <button type="button" class="kv-remove" onclick="app.removeKVItem(this)">×</button>
                        </div>
                    `).join('')}
                    <button type="button" class="kv-add" onclick="app.addKVItem('${key}')">+ Add ${title.toLowerCase()}</button>
                </div>
            </div>
        `;
    }

    renderArrayEditor(title, key, arr) {
        return `
            <div class="form-section">
                <div class="form-section-title">${title}</div>
                <div class="kv-editor" data-section="${key}">
                    ${arr.map((item, index) => `
                        <div class="kv-item">
                            <input type="text" value="${this.escapeHtml(String(item))}" placeholder="Value" data-type="value" style="flex: 2;">
                            <button type="button" class="kv-remove" onclick="app.removeKVItem(this)">×</button>
                        </div>
                    `).join('')}
                    <button type="button" class="kv-add" onclick="app.addArrayItem('${key}')">+ Add item</button>
                </div>
            </div>
        `;
    }

    addKVItem(sectionKey) {
        const section = document.querySelector(`[data-section="${sectionKey}"]`);
        const addButton = section.querySelector('.kv-add');
        
        const newItem = document.createElement('div');
        newItem.className = 'kv-item';
        newItem.innerHTML = `
            <input type="text" value="" placeholder="Key" data-type="key">
            <input type="text" value="" placeholder="Value" data-type="value">
            <button type="button" class="kv-remove" onclick="app.removeKVItem(this)">×</button>
        `;
        
        section.insertBefore(newItem, addButton);
    }

    addArrayItem(sectionKey) {
        const section = document.querySelector(`[data-section="${sectionKey}"]`);
        const addButton = section.querySelector('.kv-add');
        
        const newItem = document.createElement('div');
        newItem.className = 'kv-item';
        newItem.innerHTML = `
            <input type="text" value="" placeholder="Value" data-type="value" style="flex: 2;">
            <button type="button" class="kv-remove" onclick="app.removeKVItem(this)">×</button>
        `;
        
        section.insertBefore(newItem, addButton);
    }

    removeKVItem(button) {
        button.parentElement.remove();
    }

    // Profile Management Functions
    
    validateProfileName(name) {
        if (!name) {
            return "Profile name cannot be empty";
        }
        
        if (name.length > 255) {
            return "Profile name must be 255 characters or less";
        }
        
        // Check for valid characters only
        const validName = /^[a-zA-Z0-9_-]+$/;
        if (!validName.test(name)) {
            return "Profile name can only contain letters, numbers, hyphens, and underscores";
        }
        
        return null; // Valid
    }

    async saveProfileChanges(profileName) {
        try {
            const nameInput = document.getElementById('profile-name-input');
            const newName = nameInput ? nameInput.value.trim() : profileName;
            const formData = this.collectFormData();
            
            // Check if name changed
            if (newName !== profileName) {
                // Validate new name
                const validationError = this.validateProfileName(newName);
                if (validationError) {
                    this.showError(validationError);
                    if (nameInput) nameInput.focus();
                    return;
                }
                
                // Check if new name already exists
                if (this.profiles.some(p => p.name === newName)) {
                    this.showError(`Profile "${newName}" already exists`);
                    if (nameInput) nameInput.focus();
                    return;
                }
                
                // First update the profile content
                await this.apiCall(`/api/profiles/${encodeURIComponent(profileName)}`, {
                    method: 'PUT',
                    body: JSON.stringify(formData)
                });
                
                // Then rename the profile
                await this.apiCall(`/api/profiles/${encodeURIComponent(profileName)}/move`, {
                    method: 'POST',
                    body: JSON.stringify({ new_name: newName })
                });
                
                this.showSuccess(`Profile renamed to "${newName}" and updated successfully`);
            } else {
                // Only update content
                await this.apiCall(`/api/profiles/${encodeURIComponent(profileName)}`, {
                    method: 'PUT',
                    body: JSON.stringify(formData)
                });
                
                this.showSuccess(`Profile "${profileName}" updated successfully`);
            }
            
            this.closeModal();
            await this.loadData();
            this.renderProfiles();
        } catch (error) {
            this.showError(`Failed to save changes: ${error.message}`);
        }
    }

    collectFormData() {
        // Check current edit mode
        if (window.currentEditMode === 'raw') {
            return this.collectRawJSONData();
        } else {
            return this.collectFormFieldData();
        }
    }

    collectFormFieldData() {
        const form = document.getElementById('profile-edit-form');
        const sections = form.querySelectorAll('[data-section]');
        
        // Start with the original profile data to preserve existing fields
        const originalData = window.currentProfileData?.content || {};
        const data = {
            env: originalData.env || {},
            permissions: {
                allow: originalData.permissions?.allow || [],
                deny: originalData.permissions?.deny || []
            },
            statusLine: originalData.statusLine || {}
        };
        
        sections.forEach(section => {
            const sectionKey = section.getAttribute('data-section');
            const items = section.querySelectorAll('.kv-item');
            
            if (sectionKey === 'permissions_allow') {
                data.permissions.allow = Array.from(items).map(item => 
                    item.querySelector('[data-type="value"]').value
                ).filter(v => v.trim() !== '');
            } else if (sectionKey === 'permissions_deny') {
                data.permissions.deny = Array.from(items).map(item => 
                    item.querySelector('[data-type="value"]').value
                ).filter(v => v.trim() !== '');
            } else {
                const obj = {};
                items.forEach(item => {
                    const key = item.querySelector('[data-type="key"]')?.value;
                    const value = item.querySelector('[data-type="value"]')?.value;
                    if (key && key.trim() !== '') {
                        obj[key] = value;
                    }
                });
                
                // Always update the section based on form data
                // This ensures user modifications (including clearing fields) are respected
                data[sectionKey] = obj;
            }
        });
        
        return data;
    }

    collectRawJSONData() {
        const textarea = document.getElementById('raw-json-textarea');
        const jsonText = textarea.value.trim();
        
        try {
            const parsedData = JSON.parse(jsonText);
            
            // Validate that it has the expected structure
            if (typeof parsedData !== 'object' || parsedData === null) {
                throw new Error('Configuration must be a JSON object');
            }
            
            // Clear any previous validation errors
            this.clearJSONValidation();
            
            return parsedData;
        } catch (error) {
            this.showJSONValidationError(error.message);
            throw new Error(`Invalid JSON: ${error.message}`);
        }
    }

    syntaxHighlight(json) {
        return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
            let cls = 'number';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'key';
                } else {
                    cls = 'string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'boolean';
            } else if (/null/.test(match)) {
                cls = 'null';
            }
            return '<span class="' + cls + '">' + match + '</span>';
        });
    }

    exportConfigs() {
        this.showExportModal();
    }

    importConfigs() {
        this.showImportModal();
    }

    showExportModal() {
        const content = `
            <form id="export-form">
                <div class="form-group">
                    <label class="form-label">Export Type</label>
                    <div class="radio-group">
                        <label class="radio-label">
                            <input type="radio" name="export-type" value="all" checked>
                            Export All Profiles (${this.profiles.length} profiles)
                        </label>
                        <label class="radio-label">
                            <input type="radio" name="export-type" value="current" ${!this.currentProfile ? 'disabled' : ''}>
                            Export Current Profile Only (${this.currentProfile || 'None'})
                        </label>
                    </div>
                </div>
                
                <div class="form-group">
                    <label class="form-label">
                        <input type="checkbox" id="encrypt-export" style="margin-right: 0.5rem;">
                        Encrypt export file (recommended)
                    </label>
                </div>
                
                <div class="form-group" id="password-section" style="display: none;">
                    <label class="form-label">Encryption Password</label>
                    <input type="password" id="export-password" class="form-input" placeholder="Enter a strong password">
                    <input type="password" id="export-password-confirm" class="form-input" placeholder="Confirm password" style="margin-top: 0.5rem;">
                    <small style="color: var(--text-secondary); display: block; margin-top: 0.25rem;">
                        Use a strong password. This cannot be recovered if lost.
                    </small>
                </div>
            </form>
        `;
        
        const modal = this.createModal('Export Configurations', content, [
            { text: 'Cancel', class: 'btn-secondary', onclick: () => this.closeModal() },
            { text: 'Export & Download', class: 'btn-primary', onclick: () => this.performExport() }
        ]);
        
        document.body.appendChild(modal);
        
        // Setup event listeners for the form
        const encryptCheckbox = document.getElementById('encrypt-export');
        const passwordSection = document.getElementById('password-section');
        
        encryptCheckbox.addEventListener('change', () => {
            passwordSection.style.display = encryptCheckbox.checked ? 'block' : 'none';
        });
    }

    // Create profile functionality
    createProfile() {
        this.showCreateModal();
    }

    showCreateModal() {
        const templates = this.templates.length > 0 ? this.templates : ['default'];
        
        const content = `
            <form id="create-profile-form">
                <div class="form-group">
                    <label class="form-label">Configuration Name *</label>
                    <input type="text" id="profile-name" class="form-input" placeholder="e.g., production, development" required>
                    <small style="color: var(--text-secondary); display: block; margin-top: 0.25rem;">
                        Use descriptive names without spaces or special characters
                    </small>
                </div>
                
                <div class="form-group">
                    <label class="form-label">Template</label>
                    <select id="profile-template" class="form-input">
                        ${templates.map(t => `<option value="${t}">${t}</option>`).join('')}
                    </select>
                    <small style="color: var(--text-secondary); display: block; margin-top: 0.25rem;">
                        Choose a template to start with pre-configured settings
                    </small>
                </div>
                
                <div class="form-section">
                    <div class="form-section-title">Initial Configuration (Optional)</div>
                    
                    <div class="form-group">
                        <label class="form-label">API Token</label>
                        <input type="text" id="api-token" class="form-input" placeholder="sk-...">
                    </div>
                    
                    <div class="form-group">
                        <label class="form-label">API Base URL</label>
                        <input type="text" id="api-url" class="form-input" placeholder="https://api.example.com">
                    </div>
                </div>
            </form>
        `;
        
        const modal = this.createModal('Create New Configuration', content, [
            { text: 'Cancel', class: 'btn-secondary', onclick: () => this.closeModal() },
            { text: 'Create', class: 'btn-primary', onclick: () => this.submitCreateProfile() }
        ]);
        
        document.body.appendChild(modal);
        
        // Focus on name input
        setTimeout(() => {
            document.getElementById('profile-name')?.focus();
        }, 100);
    }

    async submitCreateProfile() {
        const nameInput = document.getElementById('profile-name');
        const templateInput = document.getElementById('profile-template');
        const tokenInput = document.getElementById('api-token');
        const urlInput = document.getElementById('api-url');
        
        const name = nameInput?.value.trim();
        const template = templateInput?.value || 'default';
        const token = tokenInput?.value.trim();
        const url = urlInput?.value.trim();
        
        if (!name) {
            this.showError('Profile name is required');
            nameInput?.focus();
            return;
        }
        
        // Validate name format (no spaces or special characters)
        if (!/^[a-zA-Z0-9-_]+$/.test(name)) {
            this.showError('Profile name can only contain letters, numbers, hyphens, and underscores');
            nameInput?.focus();
            return;
        }
        
        try {
            const requestBody = {
                name: name,
                template: template
            };
            
            // If custom values provided, include them
            if (token || url) {
                requestBody.content = {
                    env: {},
                    permissions: { allow: [], deny: [] },
                    statusLine: {}
                };
                
                if (token) {
                    requestBody.content.env.ANTHROPIC_AUTH_TOKEN = token;
                }
                if (url) {
                    requestBody.content.env.ANTHROPIC_BASE_URL = url;
                }
            }
            
            const response = await this.apiCall('/api/profiles', {
                method: 'POST',
                body: JSON.stringify(requestBody)
            });
            
            this.showSuccess(`Configuration "${name}" created successfully!`);
            this.closeModal();
            
            // Reload profiles and switch to the new one
            await this.loadData();
            this.renderProfiles();
            
            // Optional: auto-switch to the new profile
            const shouldSwitch = await this.showConfirm(
                `Switch to the new configuration "${name}"?`,
                { title: 'Switch Configuration', type: 'info', confirmText: 'Switch' }
            );
            if (shouldSwitch) {
                await this.switchProfile(name);
            }
            
        } catch (error) {
            this.showError(`Failed to create profile: ${error.message}`);
        }
    }

    // API Helper
    async apiCall(endpoint, options = {}) {
        // Set up timeout controller (60 seconds for API tests)
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 60000); // 60 seconds

        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            signal: controller.signal,
            ...options
        };

        try {
            const response = await fetch(endpoint, config);
            clearTimeout(timeoutId);
            const data = await response.json();

            if (!response.ok || !data.success) {
                throw new Error(data.error || `HTTP ${response.status}`);
            }

            return data;
        } catch (error) {
            clearTimeout(timeoutId);

            // Handle timeout specifically
            if (error.name === 'AbortError') {
                throw new Error('Request timed out after 60 seconds');
            }

            throw error;
        }
    }

    // Utility Methods
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // ==================== Toast Notification System ====================
    
    initToastContainer() {
        if (this.toastContainer) return;
        
        this.toastContainer = document.createElement('div');
        this.toastContainer.className = 'toast-container';
        document.body.appendChild(this.toastContainer);
    }

    showToast(message, type = 'info', options = {}) {
        this.initToastContainer();
        
        const {
            title = this.getToastTitle(type),
            duration = 4000,
            closable = true,
            icon = this.getToastIcon(type)
        } = options;

        const id = ++this.toastId;
        
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.dataset.toastId = id;
        
        toast.innerHTML = `
            <div class="toast-icon">${icon}</div>
            <div class="toast-content">
                <div class="toast-title">${this.escapeHtml(title)}</div>
                <div class="toast-message">${this.escapeHtml(message)}</div>
            </div>
            ${closable ? '<button class="toast-close" aria-label="Close">×</button>' : ''}
            ${duration > 0 ? `<div class="toast-progress" style="animation-duration: ${duration}ms"></div>` : ''}
        `;
        
        // Setup close button
        if (closable) {
            const closeBtn = toast.querySelector('.toast-close');
            closeBtn.addEventListener('click', () => this.removeToast(id));
        }
        
        this.toastContainer.appendChild(toast);
        this.toasts.set(id, toast);
        
        // Auto-remove after duration
        if (duration > 0) {
            setTimeout(() => this.removeToast(id), duration);
        }
        
        return id;
    }

    removeToast(id) {
        const toast = this.toasts.get(id);
        if (!toast) return;
        
        toast.classList.add('toast-removing');
        
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
            this.toasts.delete(id);
        }, 300);
    }

    getToastTitle(type) {
        const titles = {
            success: 'Success',
            error: 'Error',
            warning: 'Warning',
            info: 'Info',
            loading: 'Loading'
        };
        return titles[type] || 'Notice';
    }

    getToastIcon(type) {
        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️',
            info: 'ℹ️',
            loading: '⏳'
        };
        return icons[type] || 'ℹ️';
    }

    showSuccess(message, options = {}) {
        return this.showToast(message, 'success', options);
    }

    showError(message, options = {}) {
        return this.showToast(message, 'error', { duration: 6000, ...options });
    }

    showWarning(message, options = {}) {
        return this.showToast(message, 'warning', options);
    }

    showInfo(message, options = {}) {
        return this.showToast(message, 'info', options);
    }

    showLoading(message, options = {}) {
        return this.showToast(message, 'loading', { duration: 0, closable: false, ...options });
    }

    // ==================== Custom Dialog System ====================
    
    showConfirm(message, options = {}) {
        return new Promise((resolve) => {
            const {
                title = 'Confirm',
                confirmText = 'Confirm',
                cancelText = 'Cancel',
                type = 'warning', // 'warning', 'danger', 'info'
                confirmClass = type === 'danger' ? 'btn-danger' : 'btn-primary'
            } = options;

            const overlay = document.createElement('div');
            overlay.className = 'dialog-overlay';
            
            const dialogClass = type ? `dialog-${type}` : '';
            
            overlay.innerHTML = `
                <div class="dialog-box ${dialogClass}">
                    <div class="dialog-header">
                        <h3>${this.escapeHtml(title)}</h3>
                    </div>
                    <div class="dialog-body">
                        <p class="dialog-message">${this.escapeHtml(message)}</p>
                    </div>
                    <div class="dialog-footer">
                        <button class="btn btn-secondary dialog-cancel">${this.escapeHtml(cancelText)}</button>
                        <button class="btn ${confirmClass} dialog-confirm">${this.escapeHtml(confirmText)}</button>
                    </div>
                </div>
            `;

            // Keyboard handling - defined before closeDialog so it can be referenced
            const handleKeydown = (e) => {
                if (e.key === 'Escape') {
                    closeDialog(false);
                } else if (e.key === 'Enter') {
                    closeDialog(true);
                }
            };

            const closeDialog = (result) => {
                document.removeEventListener('keydown', handleKeydown);
                overlay.remove();
                resolve(result);
            };

            // Event listeners
            overlay.querySelector('.dialog-cancel').addEventListener('click', () => closeDialog(false));
            overlay.querySelector('.dialog-confirm').addEventListener('click', () => closeDialog(true));
            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) closeDialog(false);
            });

            document.addEventListener('keydown', handleKeydown);

            document.body.appendChild(overlay);
            overlay.querySelector('.dialog-confirm').focus();
        });
    }

    showPrompt(message, options = {}) {
        return new Promise((resolve) => {
            const {
                title = 'Input',
                defaultValue = '',
                placeholder = '',
                confirmText = 'OK',
                cancelText = 'Cancel',
                type = 'info',
                validation = null // Optional validation function
            } = options;

            const overlay = document.createElement('div');
            overlay.className = 'dialog-overlay';
            
            const dialogClass = type ? `dialog-${type}` : '';
            
            overlay.innerHTML = `
                <div class="dialog-box ${dialogClass}">
                    <div class="dialog-header">
                        <h3>${this.escapeHtml(title)}</h3>
                    </div>
                    <div class="dialog-body">
                        <p class="dialog-message">${this.escapeHtml(message)}</p>
                        <input type="text" class="dialog-input" 
                               value="${this.escapeHtml(defaultValue)}" 
                               placeholder="${this.escapeHtml(placeholder)}">
                        <div class="dialog-validation-error" style="color: var(--danger-color); font-size: 0.85rem; margin-top: 0.5rem; display: none;"></div>
                    </div>
                    <div class="dialog-footer">
                        <button class="btn btn-secondary dialog-cancel">${this.escapeHtml(cancelText)}</button>
                        <button class="btn btn-primary dialog-confirm">${this.escapeHtml(confirmText)}</button>
                    </div>
                </div>
            `;

            const input = overlay.querySelector('.dialog-input');
            const errorDiv = overlay.querySelector('.dialog-validation-error');
            const confirmBtn = overlay.querySelector('.dialog-confirm');

            // Keyboard handling - defined before closeDialog so it can be referenced
            const handleKeydown = (e) => {
                if (e.key === 'Escape') {
                    closeDialog(null);
                }
            };

            const closeDialog = (result) => {
                document.removeEventListener('keydown', handleKeydown);
                overlay.remove();
                resolve(result);
            };

            const validateAndConfirm = () => {
                const value = input.value.trim();
                
                if (validation) {
                    const error = validation(value);
                    if (error) {
                        errorDiv.textContent = error;
                        errorDiv.style.display = 'block';
                        input.focus();
                        return;
                    }
                }
                
                closeDialog(value || null);
            };

            // Event listeners
            overlay.querySelector('.dialog-cancel').addEventListener('click', () => closeDialog(null));
            confirmBtn.addEventListener('click', validateAndConfirm);
            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) closeDialog(null);
            });

            // Clear error on input
            input.addEventListener('input', () => {
                errorDiv.style.display = 'none';
            });

            // Enter key handling on input (needs to validate before closing)
            input.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    validateAndConfirm();
                }
            });

            // Document-level Escape key handling (works regardless of focus)
            document.addEventListener('keydown', handleKeydown);

            document.body.appendChild(overlay);
            input.focus();
            input.select();
        });
    }

    setupEventListeners() {
        // Handle keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'r':
                        e.preventDefault();
                        this.loadData();
                        break;
                }
            }
        });
    }
    
    // Edit Mode Toggle Functions
    toggleEditMode(mode) {
        window.currentEditMode = mode;
        
        const formEditor = document.getElementById('form-editor');
        const rawEditor = document.getElementById('raw-editor');
        
        if (mode === 'raw') {
            // Switch to raw mode
            formEditor.style.display = 'none';
            rawEditor.style.display = 'block';
            
            // Sync data from form to raw JSON
            this.syncFormToRaw();
            
            // Setup syntax highlighting overlay
            this.setupRawEditorHighlighting();
        } else {
            // Switch to form mode  
            formEditor.style.display = 'block';
            rawEditor.style.display = 'none';
        }
    }

    setupRawEditorHighlighting() {
        const textarea = document.getElementById('raw-json-textarea');
        const display = document.getElementById('raw-json-display');
        
        if (!textarea || !display) return;
        
        // Update highlighting when content changes
        const updateHighlighting = () => {
            const content = textarea.value;
            try {
                // Try to parse and format JSON
                const parsed = JSON.parse(content);
                const formatted = JSON.stringify(parsed, null, 2);
                display.innerHTML = this.syntaxHighlight(formatted);
                
                // Update textarea with formatted content
                textarea.value = formatted;
                this.clearJSONValidation();
            } catch (error) {
                // If JSON is invalid, still show it but without highlighting
                display.textContent = content;
                this.showJSONValidationError(error.message);
            }
        };
        
        // Initial highlighting
        updateHighlighting();
        
        // Update on input with debounce
        let timeout;
        textarea.addEventListener('input', () => {
            clearTimeout(timeout);
            timeout = setTimeout(updateHighlighting, 300);
        });
        
        // Sync scroll position
        textarea.addEventListener('scroll', () => {
            display.scrollTop = textarea.scrollTop;
            display.scrollLeft = textarea.scrollLeft;
        });
        
        // Focus the textarea
        textarea.focus();
    }

    syncFormToRaw() {
        try {
            const formData = this.collectFormFieldData();
            const textarea = document.getElementById('raw-json-textarea');
            const display = document.getElementById('raw-json-display');
            
            if (textarea && display) {
                const formatted = JSON.stringify(formData, null, 2);
                textarea.value = formatted;
                display.innerHTML = this.syntaxHighlight(formatted);
                this.clearJSONValidation();
            }
        } catch (error) {
            console.error('Failed to sync form to raw:', error);
        }
    }

    syncRawToForm() {
        try {
            const rawData = this.collectRawJSONData();
            // For now, we don't automatically sync back to form to avoid data loss
            // User needs to switch modes manually if they want form editing
        } catch (error) {
            // If raw JSON is invalid, don't sync
            console.log('Raw JSON invalid, not syncing to form');
        }
    }

    // JSON Validation UI Functions
    showJSONValidationError(message) {
        const validationDiv = document.getElementById('json-validation-message');
        if (validationDiv) {
            validationDiv.style.display = 'block';
            validationDiv.innerHTML = `<div style="color: #e74c3c; font-size: 14px;">❌ ${this.escapeHtml(message)}</div>`;
        }
    }

    clearJSONValidation() {
        const validationDiv = document.getElementById('json-validation-message');
        if (validationDiv) {
            validationDiv.style.display = 'none';
            validationDiv.innerHTML = '';
        }
    }

    async performExport() {
        const exportType = document.querySelector('input[name="export-type"]:checked').value;
        const encrypt = document.getElementById('encrypt-export').checked;
        
        let password = '';
        if (encrypt) {
            const password1 = document.getElementById('export-password').value;
            const password2 = document.getElementById('export-password-confirm').value;
            
            if (!password1) {
                this.showError('Password is required for encryption');
                document.getElementById('export-password').focus();
                return;
            }
            
            if (password1 !== password2) {
                this.showError('Passwords do not match');
                document.getElementById('export-password-confirm').focus();
                return;
            }
            
            if (password1.length < 8) {
                this.showError('Password must be at least 8 characters long');
                document.getElementById('export-password').focus();
                return;
            }
            
            password = password1;
        }
        
        const requestData = {
            type: exportType,
            password: password
        };
        
        if (exportType === 'current') {
            requestData.profile_name = this.currentProfile;
        }
        
        try {
            // Show loading state
            const exportButton = document.querySelector('.modal-footer .btn-primary');
            const originalText = exportButton.textContent;
            exportButton.disabled = true;
            exportButton.innerHTML = '<div class="spinner"></div>Exporting...';
            
            const response = await fetch('/api/export', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestData)
            });
            
            if (response.ok) {
                // Download file
                const blob = await response.blob();
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                
                const timestamp = new Date().toISOString().slice(0, 19).replace(/[:-]/g, '');
                const filename = `cc-switch-${exportType}-${timestamp}.ccx`;
                a.download = filename;
                
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);
                
                this.showSuccess(`Export completed successfully! File: ${filename}`);
                this.closeModal();
            } else {
                const error = await response.json();
                this.showError(`Export failed: ${error.error || 'Unknown error'}`);
            }
        } catch (error) {
            this.showError(`Export failed: ${error.message}`);
        }
    }

    // Template Management Functions
    
    validateTemplateName(name) {
        if (!name) {
            return "Template name cannot be empty";
        }
        
        if (name.length > 255) {
            return "Template name must be 255 characters or less";
        }
        
        // Check for forbidden characters
        const forbidden = /[\/\\\.\.]/;
        if (forbidden.test(name)) {
            return "Template name contains forbidden characters (/, \\, ..)";
        }
        
        // Check for valid characters only
        const validName = /^[a-zA-Z0-9_-]+$/;
        if (!validName.test(name)) {
            return "Template name can only contain letters, numbers, hyphens, and underscores";
        }
        
        // Check for reserved names
        if (name === 'default') {
            return "Cannot use reserved name 'default'";
        }
        
        return null; // Valid
    }
    
    async createTemplate() {
        const templateName = await this.showPrompt('Enter template name:', {
            title: 'Create Template',
            placeholder: 'e.g., my-template',
            validation: (value) => this.validateTemplateName(value)
        });
        
        if (!templateName) return;

        try {
            await this.apiCall('/api/templates', {
                method: 'POST',
                body: JSON.stringify({ name: templateName })
            });
            
            this.showSuccess(`Template '${templateName}' created successfully`);
            await this.loadTemplates();
            this.renderTemplates();
        } catch (error) {
            this.showError(`Failed to create template: ${error.message}`);
        }
    }

    async viewTemplate(templateName) {
        try {
            const response = await this.apiCall(`/api/templates/${encodeURIComponent(templateName)}`);
            const template = response.data;
            
            this.showModal('View Template', `
                <div class="template-view">
                    <h3>Template: ${this.escapeHtml(templateName)}</h3>
                    <p><strong>Path:</strong> ${this.escapeHtml(template.path)}</p>
                    <div class="form-group">
                        <label class="form-label">Content</label>
                        <pre class="json-display">${this.syntaxHighlight(JSON.stringify(template.content, null, 2))}</pre>
                    </div>
                </div>
            `, 'Close', null);
        } catch (error) {
            this.showError(`Failed to view template: ${error.message}`);
        }
    }

    async editTemplate(templateName) {
        try {
            const response = await this.apiCall(`/api/templates/${encodeURIComponent(templateName)}`);
            const template = response.data;
            
            const content = `
                <form id="edit-template-form">
                    <div class="form-group">
                        <label class="form-label">Template Name</label>
                        <input type="text" id="template-name-input" value="${this.escapeHtml(templateName)}" class="form-input" ${templateName === 'default' ? 'disabled' : ''}>
                        ${templateName === 'default' ? '<small style="color: var(--text-secondary);">System default template cannot be renamed</small>' : '<small style="color: var(--text-secondary);">Enter new name to rename template</small>'}
                    </div>
                    <div class="form-group">
                        <label class="form-label">Template Content (JSON)</label>
                        <textarea id="template-content" class="form-input" rows="15" style="font-family: monospace;">${JSON.stringify(template.content, null, 2)}</textarea>
                    </div>
                </form>
            `;
            
            this.showModal(`Edit Template: ${templateName}`, content, 'Save Changes', async () => {
                const nameInput = document.getElementById('template-name-input');
                const contentTextarea = document.getElementById('template-content');
                const newName = nameInput.value.trim();
                
                try {
                    // Parse and validate JSON content
                    const newContent = JSON.parse(contentTextarea.value);
                    
                    // Check if name changed and template is not default
                    if (newName !== templateName && templateName !== 'default') {
                        // Validate new name
                        const validationError = this.validateTemplateName(newName);
                        if (validationError) {
                            this.showError(validationError);
                            nameInput.focus();
                            return false;
                        }
                        
                        // Rename and update content
                        await this.apiCall(`/api/templates/${encodeURIComponent(templateName)}/move`, {
                            method: 'POST',
                            body: JSON.stringify({ new_name: newName })
                        });
                        
                        // Update content with new name
                        await this.apiCall(`/api/templates/${encodeURIComponent(newName)}`, {
                            method: 'PUT',
                            body: JSON.stringify(newContent)
                        });
                        
                        this.showSuccess(`Template renamed to '${newName}' and updated successfully`);
                    } else {
                        // Only update content
                        await this.apiCall(`/api/templates/${encodeURIComponent(templateName)}`, {
                            method: 'PUT',
                            body: JSON.stringify(newContent)
                        });
                        
                        this.showSuccess(`Template '${templateName}' updated successfully`);
                    }
                    
                    await this.loadTemplates();
                    this.renderTemplates();
                    return true;
                } catch (error) {
                    // Handle both JSON parsing errors and API call errors
                    if (error instanceof SyntaxError) {
                        this.showError(`Invalid JSON: ${error.message}`);
                        contentTextarea.focus();
                    } else {
                        this.showError(`Failed to update template: ${error.message}`);
                    }
                    return false;
                }
            });
        } catch (error) {
            this.showError(`Failed to load template for editing: ${error.message}`);
        }
    }

    async copyTemplate(templateName) {
        const newName = await this.showPrompt(`Enter name for copy of '${templateName}':`, {
            title: 'Copy Template',
            placeholder: 'e.g., my-template-copy',
            defaultValue: `${templateName}-copy`,
            validation: (value) => this.validateTemplateName(value)
        });
        
        if (!newName) return;

        try {
            await this.apiCall(`/api/templates/${encodeURIComponent(templateName)}/copy`, {
                method: 'POST',
                body: JSON.stringify({ dest_name: newName })
            });
            
            this.showSuccess(`Template '${templateName}' copied to '${newName}' successfully`);
            await this.loadTemplates();
            this.renderTemplates();
        } catch (error) {
            this.showError(`Failed to copy template: ${error.message}`);
        }
    }


    async deleteTemplate(templateName) {
        const confirmed = await this.showConfirm(
            `Are you sure you want to delete template "${templateName}"?`,
            {
                title: 'Delete Template',
                type: 'danger',
                confirmText: 'Delete',
                confirmClass: 'btn-danger'
            }
        );
        
        if (!confirmed) return;

        try {
            await this.apiCall(`/api/templates/${encodeURIComponent(templateName)}`, {
                method: 'DELETE'
            });
            
            this.showSuccess(`Template "${templateName}" deleted successfully`);
            await this.loadTemplates();
            this.renderTemplates();
        } catch (error) {
            this.showError(`Failed to delete template: ${error.message}`);
        }
    }

    async createProfileFromTemplate(templateName) {
        const profileName = await this.showPrompt(
            `Enter name for new configuration from template '${templateName}':`,
            {
                title: 'Create Configuration',
                placeholder: 'e.g., my-config',
                validation: (value) => this.validateProfileName(value)
            }
        );
        
        if (!profileName) return;

        try {
            await this.apiCall(`/api/templates/${encodeURIComponent(templateName)}/copy`, {
                method: 'POST',
                body: JSON.stringify({ 
                    dest_name: profileName,
                    to_config: true 
                })
            });
            
            this.showSuccess(`Configuration '${profileName}' created from template '${templateName}' successfully`);
            await this.loadData();
            this.showSection('profiles');
        } catch (error) {
            this.showError(`Failed to create configuration from template: ${error.message}`);
        }
    }

    showImportModal() {
        const content = `
            <form id="import-form">
                <div class="form-group">
                    <label class="form-label">Select CCX File</label>
                    <div class="file-upload-area" onclick="document.getElementById('import-file').click()" style="border: 2px dashed #ccc; padding: 2rem; text-align: center; cursor: pointer; border-radius: 8px;">
                        <input type="file" id="import-file" accept=".ccx" style="display: none;">
                        <div class="file-upload-content">
                            <div class="file-upload-icon" style="font-size: 2rem; margin-bottom: 1rem;">📁</div>
                            <div class="file-upload-text">Click to select a CCX file or drag & drop</div>
                            <div class="file-upload-hint" style="color: var(--text-secondary); margin-top: 0.5rem;">Supported format: .ccx files exported from cc-switch</div>
                        </div>
                    </div>
                    <div id="file-info" style="display: none; margin-top: 1rem; padding: 1rem; background: var(--bg-secondary); border-radius: 4px;">
                        <div class="file-info-item">
                            <strong>File:</strong> <span id="file-name"></span>
                        </div>
                        <div class="file-info-item">
                            <strong>Size:</strong> <span id="file-size"></span>
                        </div>
                    </div>
                </div>
                
                <div class="form-group" id="password-section-import" style="display: none;">
                    <label class="form-label">Decryption Password</label>
                    <input type="password" id="import-password" class="form-input" placeholder="Enter password to decrypt file">
                    <small style="color: var(--text-secondary); display: block; margin-top: 0.25rem;">
                        This file appears to be encrypted. Enter the password used during export.
                    </small>
                </div>
                
                <div class="form-group">
                    <label class="form-label">Import Options</label>
                    <div class="checkbox-group">
                        <label class="checkbox-label" style="display: block; margin-bottom: 0.5rem;">
                            <input type="checkbox" id="import-preview" style="margin-right: 0.5rem;">
                            Preview only (don't actually import)
                        </label>
                    </div>
                    
                    <div class="form-group" style="margin-top: 1rem;">
                        <label class="form-label">Conflict Resolution</label>
                        <select id="import-conflict" class="form-input">
                            <option value="both">Rename conflicting profiles (recommended)</option>
                            <option value="skip">Skip conflicting profiles</option>
                            <option value="overwrite">Overwrite existing profiles</option>
                        </select>
                        <small style="color: var(--text-secondary); display: block; margin-top: 0.25rem;">
                            Choose how to handle profiles with names that already exist
                        </small>
                    </div>
                </div>
            </form>
        `;
        
        const modal = this.createModal('Import Configurations', content, [
            { text: 'Cancel', class: 'btn-secondary', onclick: () => this.closeModal() },
            { text: 'Import', class: 'btn-primary', onclick: () => this.performImport(), id: 'import-button', disabled: true }
        ]);
        
        document.body.appendChild(modal);
        
        // Setup file input handling
        this.setupImportFileHandling();
    }

    setupImportFileHandling() {
        const fileInput = document.getElementById('import-file');
        const fileInfo = document.getElementById('file-info');
        const fileName = document.getElementById('file-name');
        const fileSize = document.getElementById('file-size');
        const importButton = document.getElementById('import-button');
        const passwordSection = document.getElementById('password-section-import');
        
        fileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                fileName.textContent = file.name;
                fileSize.textContent = this.formatFileSize(file.size);
                fileInfo.style.display = 'block';
                importButton.disabled = false;
                
                // For now, assume all .ccx files might be encrypted
                if (file.name.endsWith('.ccx')) {
                    passwordSection.style.display = 'block';
                }
            } else {
                fileInfo.style.display = 'none';
                importButton.disabled = true;
                passwordSection.style.display = 'none';
            }
        });
        
        // Setup drag & drop
        const uploadArea = document.querySelector('.file-upload-area');
        
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.style.borderColor = '#007bff';
            uploadArea.style.backgroundColor = '#f8f9fa';
        });
        
        uploadArea.addEventListener('dragleave', (e) => {
            uploadArea.style.borderColor = '#ccc';
            uploadArea.style.backgroundColor = 'transparent';
        });
        
        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.style.borderColor = '#ccc';
            uploadArea.style.backgroundColor = 'transparent';
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                fileInput.files = files;
                fileInput.dispatchEvent(new Event('change'));
            }
        });
    }

    async performImport() {
        const fileInput = document.getElementById('import-file');
        const file = fileInput.files[0];
        
        if (!file) {
            this.showError('Please select a file to import');
            return;
        }
        
        const formData = new FormData();
        formData.append('file', file);
        
        const options = {
            conflict_mode: document.getElementById('import-conflict').value,
            dry_run: document.getElementById('import-preview').checked
        };
        
        const password = document.getElementById('import-password').value;
        if (password) {
            formData.append('password', password);
        }
        
        formData.append('options', JSON.stringify(options));
        
        try {
            // Show loading state
            const importButton = document.getElementById('import-button');
            const originalText = importButton.textContent;
            importButton.disabled = true;
            importButton.innerHTML = '<div class="spinner"></div>Importing...';
            
            const response = await fetch('/api/import', {
                method: 'POST',
                body: formData
            });
            
            const result = await response.json();
            
            if (response.ok && result.success) {
                this.showImportResults(result.data);
                this.closeModal();
                
                // If not a dry run, reload data
                if (!options.dry_run) {
                    await this.loadData();
                    this.renderProfiles();
                }
            } else {
                this.showError(`Import failed: ${result.error || 'Unknown error'}`);
            }
        } catch (error) {
            this.showError(`Import failed: ${error.message}`);
        }
    }

    showImportResults(result) {
        const isDryRun = result.dry_run || false;
        const title = isDryRun ? 'Import Preview Results' : 'Import Results';
        
        let content = `
            <div class="import-results">
                <div class="result-summary" style="margin-bottom: 1.5rem;">
                    <h3>Summary</h3>
                    <div class="summary-item" style="margin-bottom: 0.5rem;">
                        <strong>Total profiles in file:</strong> ${result.total_profiles || 0}
                    </div>
                    <div class="summary-item" style="margin-bottom: 0.5rem;">
                        <strong>${isDryRun ? 'Would import' : 'Imported'}:</strong> 
                        <span style="color: #28a745;">${result.imported_count || 0}</span>
                    </div>`;
        
        if (result.skipped_count > 0) {
            content += `
                    <div class="summary-item" style="margin-bottom: 0.5rem;">
                        <strong>Skipped:</strong> 
                        <span style="color: #ffc107;">${result.skipped_count}</span>
                    </div>`;
        }
        
        if (result.renamed_count > 0) {
            content += `
                    <div class="summary-item" style="margin-bottom: 0.5rem;">
                        <strong>Renamed:</strong> 
                        <span style="color: #17a2b8;">${result.renamed_count}</span>
                    </div>`;
        }
        
        if (result.error_count > 0) {
            content += `
                    <div class="summary-item" style="margin-bottom: 0.5rem;">
                        <strong>Errors:</strong> 
                        <span style="color: #dc3545;">${result.error_count}</span>
                    </div>`;
        }
        
        content += '</div>';
        
        if (result.profiles_imported && result.profiles_imported.length > 0) {
            content += `
                <div class="result-section" style="margin-bottom: 1.5rem;">
                    <h4>${isDryRun ? 'Profiles that would be imported:' : 'Successfully imported profiles:'}</h4>
                    <ul style="list-style: none; padding: 0;">`;
            
            result.profiles_imported.forEach(profile => {
                content += `<li style="margin-bottom: 0.25rem; color: #28a745;">✅ ${this.escapeHtml(profile)}</li>`;
            });
            
            content += '</ul></div>';
        }
        
        if (result.conflicts && result.conflicts.length > 0) {
            content += `
                <div class="result-section" style="margin-bottom: 1.5rem;">
                    <h4>Conflicts handled:</h4>
                    <ul style="list-style: none; padding: 0;">`;
            
            result.conflicts.forEach(conflict => {
                content += `<li style="margin-bottom: 0.25rem; color: #ffc107;">⚠️ ${this.escapeHtml(conflict)}</li>`;
            });
            
            content += '</ul></div>';
        }
        
        if (result.errors && result.errors.length > 0) {
            content += `
                <div class="result-section">
                    <h4>Errors encountered:</h4>
                    <ul style="list-style: none; padding: 0;">`;
            
            result.errors.forEach(error => {
                content += `<li style="margin-bottom: 0.25rem; color: #dc3545;">❌ ${this.escapeHtml(error)}</li>`;
            });
            
            content += '</ul></div>';
        }
        
        content += '</div>';
        
        this.showModal(title, content, 'Close', null);
    }

    formatFileSize(bytes) {
        const units = ['B', 'KB', 'MB', 'GB'];
        let size = bytes;
        let unitIndex = 0;
        
        while (size >= 1024 && unitIndex < units.length - 1) {
            size /= 1024;
            unitIndex++;
        }
        
        return `${size.toFixed(1)} ${units[unitIndex]}`;
    }

    async checkForUpdates() {
        try {
            const response = await this.apiCall('/api/version');
            const data = response.data;
            
            if (data.has_update) {
                const banner = document.getElementById('update-banner');
                const currentSpan = document.getElementById('update-current');
                const latestSpan = document.getElementById('update-latest');
                const dismissBtn = document.getElementById('update-dismiss');
                
                if (banner && currentSpan && latestSpan) {
                    currentSpan.textContent = `v${data.current_version}`;
                    latestSpan.textContent = `v${data.latest_version}`;
                    banner.style.display = 'block';
                    
                    // Setup dismiss button
                    if (dismissBtn) {
                        dismissBtn.addEventListener('click', () => {
                            banner.style.display = 'none';
                            // Store dismissal in session storage
                            sessionStorage.setItem('update-dismissed', data.latest_version);
                        });
                    }
                    
                    // Check if already dismissed this version
                    const dismissedVersion = sessionStorage.getItem('update-dismissed');
                    if (dismissedVersion === data.latest_version) {
                        banner.style.display = 'none';
                    }
                }
            }
        } catch (error) {
            // Silently ignore update check errors
            console.log('Update check skipped:', error.message);
        }
    }
}

// Initialize the application when the DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.app = new CCSwitch();
});