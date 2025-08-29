class CCSwitch {
    constructor() {
        this.currentProfile = null;
        this.profiles = [];
        this.isEmptyMode = false;
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
        if (!confirm(`Are you sure you want to delete configuration "${profileName}"?`)) {
            return;
        }

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
                    timeout: 10
                })
            });

            const result = response.data;
            
            resultsContent.innerHTML = `
                <div class="status ${result.IsConnectable ? 'status-online' : 'status-offline'}">
                    ${result.IsConnectable ? '✅ Connected' : '❌ Connection Failed'}
                </div>
                <p><strong>Profile:</strong> ${result.ProfileName}</p>
                <p><strong>Response Time:</strong> ${Math.round(result.ResponseTime / 1000000)}ms</p>
                <p><strong>Tested At:</strong> ${new Date(result.TestedAt).toLocaleString()}</p>
                ${result.Error ? `<p class="status-offline"><strong>Error:</strong> ${result.Error}</p>` : ''}
            `;
            
            resultsDiv.style.display = 'block';
        } catch (error) {
            this.showError(`Test failed: ${error.message}`);
        } finally {
            testButton.disabled = false;
            testButton.innerHTML = 'Run Test';
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
            <div class="profile-metadata">
                <dt>Name:</dt>
                <dd>${this.escapeHtml(profile.name)}</dd>
                <dt>Path:</dt>
                <dd>${this.escapeHtml(profile.path)}</dd>
            </div>
            
            <form id="profile-edit-form">
                ${this.renderEditSection('Environment Variables', 'env', content.env || {})}
                ${this.renderEditSection('Permissions - Allow', 'permissions_allow', content.permissions?.allow || [])}
                ${this.renderEditSection('Permissions - Deny', 'permissions_deny', content.permissions?.deny || [])}
                ${this.renderEditSection('Status Line', 'statusLine', content.statusLine || {})}
            </form>
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

    async saveProfileChanges(profileName) {
        try {
            const formData = this.collectFormData();
            
            // Call update API (placeholder - you'll need to implement the PUT endpoint)
            const response = await this.apiCall(`/api/profiles/${encodeURIComponent(profileName)}`, {
                method: 'PUT',
                body: JSON.stringify(formData)
            });
            
            this.showSuccess(`Profile "${profileName}" updated successfully`);
            this.closeModal();
            await this.loadData();
            this.renderProfiles();
        } catch (error) {
            this.showError(`Failed to save changes: ${error.message}`);
        }
    }

    collectFormData() {
        const form = document.getElementById('profile-edit-form');
        const sections = form.querySelectorAll('[data-section]');
        const data = { env: {}, permissions: { allow: [], deny: [] }, statusLine: {} };
        
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
                data[sectionKey] = obj;
            }
        });
        
        return data;
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
        alert('Export configurations functionality coming soon!');
    }

    importConfigs() {
        alert('Import configurations functionality coming soon!');
    }

    // Create profile functionality
    createProfile() {
        this.showCreateModal();
    }

    showCreateModal() {
        const templates = ['default']; // Could be fetched from API
        
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
            if (confirm(`Switch to the new configuration "${name}"?`)) {
                await this.switchProfile(name);
            }
            
        } catch (error) {
            this.showError(`Failed to create profile: ${error.message}`);
        }
    }

    // API Helper
    async apiCall(endpoint, options = {}) {
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        const response = await fetch(endpoint, config);
        const data = await response.json();

        if (!response.ok || !data.success) {
            throw new Error(data.error || `HTTP ${response.status}`);
        }

        return data;
    }

    // Utility Methods
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    showSuccess(message) {
        // Simple alert for now - can be replaced with proper notifications
        alert(`✅ ${message}`);
    }

    showError(message) {
        // Simple alert for now - can be replaced with proper notifications
        alert(`❌ ${message}`);
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
}

// Initialize the application when the DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.app = new CCSwitch();
});