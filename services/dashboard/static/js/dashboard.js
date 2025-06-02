// Toast functionality
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    
    const bgColor = {
        success: 'bg-green-500',
        error: 'bg-red-500',
        info: 'bg-blue-500',
        warning: 'bg-yellow-500'
    }[type] || 'bg-blue-500';
    
    toast.className = `${bgColor} text-white px-6 py-4 rounded-lg shadow-lg flex items-center space-x-3 transform transition-all duration-300 translate-x-full`;
    toast.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check' : type === 'error' ? 'exclamation' : 'info'}-circle"></i>
        <span>${message}</span>
        <button onclick="this.parentElement.remove()" class="ml-auto">
            <i class="fas fa-times"></i>
        </button>
    `;
    
    container.appendChild(toast);
    
    // Animate in
    setTimeout(() => {
        toast.classList.remove('translate-x-full');
    }, 100);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
        toast.classList.add('translate-x-full');
        setTimeout(() => toast.remove(), 300);
    }, 5000);
}

// Players data management
function playersData() {
    return {
        players: [],
        loading: true,
        showEditModal: false,
        editingPlayer: {
            uuid: '',
            name: '',
            labels: [],
            annotations: []
        },
        
        async loadPlayers() {
            this.loading = true;
            try {
                const response = await fetch('/api/players');
                this.players = await response.json();
            } catch (error) {
                console.error('Error loading players:', error);
                showToast('Failed to load players', 'error');
            } finally {
                this.loading = false;
            }
        },
        
        refreshPlayers() {
            this.loadPlayers();
            showToast('Players refreshed', 'info');
        },
        
        editPlayer(player) {
            this.editingPlayer = {
                uuid: player.uuid,
                name: player.name || '',
                labels: Object.entries(player.labels || {}).map(([key, value]) => ({key, value})),
                annotations: Object.entries(player.annotations || {}).map(([key, value]) => ({key, value}))
            };
            this.showEditModal = true;
        },
        
        async savePlayer() {
            const labels = {};
            const annotations = {};
            
            this.editingPlayer.labels.forEach(({key, value}) => {
                if (key.trim()) labels[key.trim()] = value;
            });
            
            this.editingPlayer.annotations.forEach(({key, value}) => {
                if (key.trim()) annotations[key.trim()] = value;
            });
            
            // Add player name to annotations
            if (this.editingPlayer.name) {
                annotations['player_name'] = this.editingPlayer.name;
            }
            
            try {
                const response = await fetch(`/api/players/${this.editingPlayer.uuid}/update`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ labels, annotations })
                });
                
                const result = await response.json();
                if (result.error) {
                    showToast(result.error, 'error');
                } else {
                    showToast('Player updated successfully', 'success');
                    this.showEditModal = false;
                    this.loadPlayers();
                }
            } catch (error) {
                console.error('Error updating player:', error);
                showToast('Failed to update player', 'error');
            }
        },
        
        getStatusBadgeClass(status) {
            const classes = {
                online: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800',
                offline: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800',
                unknown: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800'
            };
            return classes[status?.toLowerCase()] || classes.unknown;
        }
    };
}

// Servers data management
function serversData() {
    return {
        servers: [],
        loading: true,
        showEditModal: false,
        editingServer: {
            name: '',
            labels: [],
            annotations: []
        },
        
        async loadServers() {
            this.loading = true;
            try {
                const response = await fetch('/api/servers');
                this.servers = await response.json();
            } catch (error) {
                console.error('Error loading servers:', error);
                showToast('Failed to load servers', 'error');
            } finally {
                this.loading = false;
            }
        },
        
        refreshServers() {
            this.loadServers();
            showToast('Servers refreshed', 'info');
        },
        
        editServer(server) {
            this.editingServer = {
                name: server.name,
                labels: Object.entries(server.labels || {}).map(([key, value]) => ({key, value})),
                annotations: Object.entries(server.annotations || {}).map(([key, value]) => ({key, value}))
            };
            this.showEditModal = true;
        },
        
        async saveServer() {
            const labels = {};
            const annotations = {};
            
            this.editingServer.labels.forEach(({key, value}) => {
                if (key.trim()) labels[key.trim()] = value;
            });
            
            this.editingServer.annotations.forEach(({key, value}) => {
                if (key.trim()) annotations[key.trim()] = value;
            });
            
            try {
                const response = await fetch(`/api/servers/${this.editingServer.name}/update`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ labels, annotations })
                });
                
                const result = await response.json();
                if (result.error) {
                    showToast(result.error, 'error');
                } else {
                    showToast('Server updated successfully', 'success');
                    this.showEditModal = false;
                    this.loadServers();
                }
            } catch (error) {
                console.error('Error updating server:', error);
                showToast('Failed to update server', 'error');
            }
        },
        
        getStatusBadgeClass(status) {
            const classes = {
                online: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800',
                offline: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800',
                unknown: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800'
            };
            return classes[status?.toLowerCase()] || classes.unknown;
        }
    };
}

// Alpine.js directive for x-cloak
document.addEventListener('alpine:init', () => {
    Alpine.directive('cloak', (el) => el.removeAttribute('x-cloak'));
});
