let currentPlayers = [];

function loadPlayers() {
    fetch('/api/players')
        .then(response => response.json())
        .then(players => {
            currentPlayers = players;
            displayPlayers(players);
        })
        .catch(error => {
            console.error('Error loading players:', error);
            showToast('Failed to load players', 'danger');
        });
}

function displayPlayers(players) {
    const tbody = document.getElementById('playersTableBody');
    
    if (players.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="text-center">No players found</td></tr>';
        return;
    }
    
    tbody.innerHTML = players.map(player => `
        <tr>
            <td><code>${player.uuid}</code></td>
            <td>${player.name || 'Unknown'}</td>
            <td>${getStatusBadge(player.status)}</td>
            <td><div class="labels-display">${formatLabels(player.labels)}</div></td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="editPlayer('${player.uuid}')">
                    <i class="fas fa-edit"></i> Edit
                </button>
            </td>
        </tr>
    `).join('');
}

function editPlayer(uuid) {
    const player = currentPlayers.find(p => p.uuid === uuid);
    if (!player) return;
    
    document.getElementById('editPlayerUUID').value = uuid;
    document.getElementById('editPlayerName').value = player.name || '';
    
    populateKeyValueEditor('editPlayerLabels', player.labels || {}, 'label');
    populateKeyValueEditor('editPlayerAnnotations', player.annotations || {}, 'annotation');
    
    const modal = new bootstrap.Modal(document.getElementById('editPlayerModal'));
    modal.show();
}

function populateKeyValueEditor(containerId, data, type) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';
    
    Object.entries(data).forEach(([key, value]) => {
        addKeyValuePair(container, key, value, type);
    });
}

function addKeyValuePair(container, key = '', value = '', type) {
    const div = document.createElement('div');
    div.className = 'key-value-pair';
    div.innerHTML = `
        <input type="text" class="form-control" placeholder="Key" value="${key}">
        <input type="text" class="form-control" placeholder="Value" value="${value}">
        <button type="button" class="btn btn-outline-danger btn-sm" onclick="this.parentElement.remove()">
            <i class="fas fa-times"></i>
        </button>
    `;
    container.appendChild(div);
}

function addPlayerLabel() {
    const container = document.getElementById('editPlayerLabels');
    addKeyValuePair(container, '', '', 'label');
}

function addPlayerAnnotation() {
    const container = document.getElementById('editPlayerAnnotations');
    addKeyValuePair(container, '', '', 'annotation');
}

function savePlayer() {
    const uuid = document.getElementById('editPlayerUUID').value;
    const name = document.getElementById('editPlayerName').value;
    
    const labels = collectKeyValuePairs('editPlayerLabels');
    const annotations = collectKeyValuePairs('editPlayerAnnotations');
    
    // Add player name to annotations if provided
    if (name) {
        annotations['player_name'] = name;
    }
    
    const updateData = { labels, annotations };
    
    fetch(`/api/players/${uuid}/update`, {
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
            showToast('Player updated successfully', 'success');
            bootstrap.Modal.getInstance(document.getElementById('editPlayerModal')).hide();
            loadPlayers();
        }
    })
    .catch(error => {
        console.error('Error updating player:', error);
        showToast('Failed to update player', 'danger');
    });
}

function collectKeyValuePairs(containerId) {
    const container = document.getElementById(containerId);
    const pairs = {};
    
    container.querySelectorAll('.key-value-pair').forEach(pair => {
        const inputs = pair.querySelectorAll('input');
        const key = inputs[0].value.trim();
        const value = inputs[1].value.trim();
        
        if (key) {
            pairs[key] = value;
        }
    });
    
    return pairs;
}

function refreshPlayers() {
    loadPlayers();
    showToast('Players refreshed', 'info');
}

// Load players when page loads
document.addEventListener('DOMContentLoaded', loadPlayers);
