
// Billing Microservice Logic

async function createBillingFolder() {
    const name = document.getElementById('bill-folder-name').value;
    document.getElementById('res-bill-folder').innerText = 'Sending request...';

    const response = await fetch(`${BILLING_URL}/folders`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name })
    });

    renderResponse(response, 'res-bill-folder');
}

async function createTicket() {
    const folderId = document.getElementById('bill-ticket-folder').value;
    const total = document.getElementById('bill-ticket-total').value;
    const desc = document.getElementById('bill-ticket-desc').value;
    document.getElementById('res-bill-ticket').innerText = 'Sending request...';

    const response = await fetch(`${BILLING_URL}/tickets`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            folder_id: folderId,
            total: total.toString(),
            item_id: "REQ-" + Math.floor(Math.random() * 1000),
            description: desc
        })
    });

    renderResponse(response, 'res-bill-ticket');
    if (response.ok) searchTickets(); // Refresh list if search is active
}

async function searchTickets() {
    const folderId = document.getElementById('bill-search-folder').value.trim();
    const display = document.getElementById('tickets-list-display');
    display.innerHTML = '<div class="loading-spinner">Cargando tickets...</div>';

    let url = `${BILLING_URL}/tickets?page=0&size=50`;
    if (folderId) {
        url += `&folder_id=${folderId}`;
    }

    try {
        const response = await fetch(url);
        if (response.ok) {
            const data = await response.json();
            const tickets = data.content ? data.content : (Array.isArray(data) ? data : []);
            renderTicketsList(tickets);
        } else {
            renderResponse(response, 'tickets-list-display');
        }
    } catch (e) {
        display.innerHTML = `<div class="error-msg">Error: ${e.message}</div>`;
    }
}

function renderTicketsList(tickets) {
    const display = document.getElementById('tickets-list-display');
    if (!tickets || tickets.length === 0) {
        display.innerHTML = '<div class="p-4 bg-gray-100 rounded">No se encontraron tickets.</div>';
        return;
    }

    let html = `
    <table class="inventory-table">
        <thead>
            <tr>
                <th>ID</th>
                <th>Estado</th>
                <th>Total</th>
                <th>Descripción</th>
                <th>Acciones</th>
            </tr>
        </thead>
        <tbody>`;

    tickets.forEach(t => {
        const stateLower = (t.state || 'unknown').toLowerCase();
        let badgeClass;
        let disableCancelBtn = false;

        switch (stateLower) {
            case 'canceled':
                badgeClass = 'state-canceled';
                disableCancelBtn = true;
                break;
            case 'closed':
            case 'paid':
                badgeClass = 'state-closed';
                disableCancelBtn = true;
                break;
            case 'pending':
                badgeClass = 'state-pending';
                break;
            default:
                badgeClass = 'state-pending';
        }

        html += `<tr>
            <td title="${t.id}" style="font-family:monospace; font-size:0.8rem;">${t.id.substring(0, 8)}...</td>
            <td><span class="state-badge ${badgeClass}">${t.state || 'N/A'}</span></td>
            <td><strong>$${t.total}</strong></td>
            <td>${t.description}</td>
            <td class="ticket-actions">
                <button class="btn-small btn-view" 
                    onclick="viewTicketDetail('${t.id}')"
                >Ver</button>
                <button class="btn-small btn-cancel"
                    onclick="cancelTicket('${t.id}')"
                    ${disableCancelBtn ? 'disabled' : ''}
                >Cancelar</button>
            </td>
        </tr>`;
    });
    html += '</tbody></table>';
    display.innerHTML = html;
}

async function viewTicketDetail(id) {
    const modal = document.getElementById('ticket-modal');
    const body = document.getElementById('ticket-modal-body');
    modal.classList.add('active');
    body.innerHTML = '<div class="loading-spinner">Cargando detalle...</div>';

    try {
        const response = await fetch(`${BILLING_URL}/tickets/${id}`);
        if (response.ok) {
            const t = await response.json();

            const stateLower = (t.state || 'unknown').toLowerCase();
            let badgeClass = 'state-pending';
            if (stateLower === 'canceled') badgeClass = 'state-canceled';
            if (stateLower === 'closed' || stateLower === 'paid') badgeClass = 'state-closed';

            body.innerHTML = `
                <div style="line-height: 1.8;">
                    <div style="margin-bottom:15px; border-bottom:1px solid #eee; padding-bottom:10px;">
                        <span class="state-badge ${badgeClass}" style="font-size:1rem;">${t.state || 'Unknown'}</span>
                    </div>
                    <p><strong>ID Ticket:</strong> ${t.id}</p>
                    <p><strong>Folder ID:</strong> ${t.folder_id || (t.folder ? t.folder.id : 'N/A')}</p>
                    <p><strong>Item Relacionado:</strong> ${t.item_id || 'N/A'}</p>
                    <p><strong>Descripción:</strong> ${t.description}</p>
                    <p><strong>Creado:</strong> ${t.created_at ? new Date(t.created_at).toLocaleString() : 'N/A'}</p>
                    <p><strong>Total:</strong> <span style="font-size:1.5rem; color:var(--accent); font-weight:bold; display:block; margin-top:5px;">$${t.total}</span></p>
                </div>
            `;
        } else {
            body.innerHTML = `<div class="error-msg">Error al cargar: ${response.status}</div>`;
        }
    } catch (e) {
        body.innerHTML = `<div class="error-msg">Error de conexión</div>`;
    }
}

function closeTicketModal() {
    document.getElementById('ticket-modal').classList.remove('active');
}

async function cancelTicket(id) {
    if (!confirm("¿Estás seguro de que quieres anular este ticket? Esta acción no se puede deshacer.")) return;

    try {
        const response = await fetch(`${BILLING_URL}/tickets/${id}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            alert("Ticket anulado correctamente.");
            searchTickets();
        } else {
            alert("Error al anular ticket.");
        }
    } catch (e) {
        alert("Error de conexión al intentar anular.");
    }
}

async function getFolderById() {
    const folderId = document.getElementById('bill-folder-op-id').value;
    document.getElementById('res-bill-folder-op').innerText = 'Sending request...';

    const response = await fetch(`${BILLING_URL}/folders/${folderId}`);
    renderResponse(response, 'res-bill-folder-op');
}

async function closeFolder() {
    const folderId = document.getElementById('bill-folder-op-id').value;
    if (!confirm(`¿Cerrar carpeta ${folderId}?`)) return;

    document.getElementById('res-bill-folder-op').innerText = 'Sending request...';

    const response = await fetch(`${BILLING_URL}/folders/${folderId}/close`, {
        method: 'POST'
    });
    renderResponse(response, 'res-bill-folder-op');
}

async function makePayment() {
    const method = document.getElementById('bill-pay-method').value;
    const desc = document.getElementById('bill-pay-desc').value;
    const ticketsStr = document.getElementById('bill-pay-tickets').value;
    const tickets = ticketsStr.split(',').map(s => s.trim()).filter(s => s.length > 0);

    document.getElementById('res-bill-pay').innerText = 'Sending request...';

    const response = await fetch(`${BILLING_URL}/payments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            payment_method: method,
            description: desc,
            tickets: tickets
        })
    });
    renderResponse(response, 'res-bill-pay');
}

async function checkBillingHealth() {
    const btn = document.getElementById('bill-health-btn');
    btn.innerText = 'Verificando...';
    btn.style.backgroundColor = '#f39c12'; // Orange/Pending

    try {
        const response = await fetch(`${BILLING_URL}/health`);
        if (response) {
            btn.innerText = 'OK!';
            btn.style.backgroundColor = '#2ecc71'; // Green/Success
        } else {
            throw new Error('Status ' + response.status);
        }
    } catch (e) {
        btn.innerText = 'Error';
        btn.style.backgroundColor = '#e74c3c'; // Red/Error
    }
}
