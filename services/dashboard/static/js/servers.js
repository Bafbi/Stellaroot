let currentServers = [];

function loadServers() {
    fetch('/api/servers')
        .then(response => response.json())
        .then(servers => {
            currentServers = servers;
            displayServers(servers);
        })
        .catch(error => {
            console.error('Error loading servers:', error);
            showToast('Failed to load servers', 'danger');
        });
}

function displayServers(servers) {
    const tbody = document.getElementById('serversTableBody');
    
    if (servers.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="text-center">No servers found</td></tr>';
        return;
    }
    
    tbody.innerHTML = servers.map(server => `
        <tr>
            <td><strong>${server.name}</strong></td>
            <td>${getStatusBadge(server.status)}</td>
            <td><span class="badge bg-info">${server.player_count || 0} players</span></td>
            <td><div class="labels-display">${formatLabels(server.labels)}</div></td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="editServer('${server.name}')">
                    <i class="fas fa-edit"></i> Edit
                </button>
            </td>
        </tr>
    `).join('');
}

function editServer(name) {
    const server = currentServers.find(s => s.name === name);
    if (!server) return;
    
    document.getElementById('editServerName').value = name;
    document.getElementById('editServerDisplayName').value = name;
    
    populateKeyValueEditor('editServerLabels', server.labels || {}, 'label');
    populateKeyValueEditor('editServerAnnotations', server.annotations || {}, 'annotation');
    
    const modal = new bootstrap.Modal(document.getElementById('editServerModal'));
    modal.show();
}

function addServerLabel() {
    const container = document.getElementById('editServerLabels');
    addKeyValuePair(container, '', '', 'label');
}

function addServerAnnotation() {
    const container = document.getElementById('editServerAnnotations');
    addKeyValuePair(container, '', '', 'annotation');
}

function saveServer() {
    const name = document.getElementById('editServerName').value;
    
    const labels = collectKeyValuePairs('editServerLabels');
    const annotations = collectKeyValuePairs('editServerAnnotations');
    
    const updateData = { labels, annotations };
    
    fetch(`/api/servers/${name}/update`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(updateData),
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            showToast(data.error, 'danger');
        } else {
            showToast('Server updated successfully', 'success');
            bootstrap.Modal.getInstance(document.getElementById('editServerModal')).hide();
            loadServers();
        }
    })
    .catch(error => {
        console.error('Error updating server:', error);
        showToast('Failed to update server', 'danger');
    });
}

function refreshServers() {
    loadServers();
    showToast('Servers refreshed', 'info');
}

// Load servers when page loads
document.addEventListener('DOMContentLoaded', loadServers);
