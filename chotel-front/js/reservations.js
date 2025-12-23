
// Reservations Microservice Logic

async function createReservation() {
    const roomId = document.getElementById('res-room-id').value;
    const guestId = document.getElementById('res-guest-id').value;
    const startDate = document.getElementById('res-start-date').value + ":00.0-03:00";
    const endDate = document.getElementById('res-end-date').value + ":00.0-03:00";
    document.getElementById('res-display-create').innerText = 'Sending request...';

    const response = await fetch(`${RESERVATIONS_URL}/reservations`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            roomId: parseInt(roomId),
            guestId: guestId,
            startDate: startDate,
            endDate: endDate
        })
    });

    renderResponse(response, 'res-display-create');
}

// State for rooms pagination
let currentRoomsPage = 1;
let allRoomsData = [];
const ROOMS_PAGE_SIZE = 10;

async function searchRooms() {
    const container = document.getElementById('res-display-search');
    container.innerHTML = '<div class="loading-spinner">Buscando habitaciones...</div>';

    try {
        // Fetch all rooms (large page size)
        const response = await fetch(`${RESERVATIONS_URL}/rooms?page=0&pageSize=2000`, {
            method: 'GET'
        });

        if (response.ok) {
            const data = await response.json();
            allRoomsData = Array.isArray(data) ? data : (data.content || []);
            currentRoomsPage = 1;

            if (allRoomsData.length === 0) {
                container.innerHTML = '<div class="p-4 bg-gray-100 rounded">No se encontraron habitaciones.</div>';
                return;
            }
            renderRoomsPage();
        } else {
            container.innerHTML = `<div class="error-msg">Error: ${response.status}</div>`;
        }
    } catch (error) {
        container.innerHTML = `<div class="error-msg">Error de conexión: ${error.message}</div>`;
    }
}

function renderRoomsPage() {
    const container = document.getElementById('res-display-search');
    const start = (currentRoomsPage - 1) * ROOMS_PAGE_SIZE;
    const end = start + ROOMS_PAGE_SIZE;
    const pageItems = allRoomsData.slice(start, end);
    const totalPages = Math.ceil(allRoomsData.length / ROOMS_PAGE_SIZE);

    // Reuse existing render helper, but capturing HTML first
    // We will generate the HTML manually to include grid + pagination similar to reservations
    let html = '<div class="reservations-grid">';
    pageItems.forEach(room => {
        const isActive = room.active;
        html += `
        <div class="reservation-card ${isActive ? 'active' : 'inactive'}" style="border-left: 4px solid ${isActive ? '#2ecc71' : '#e74c3c'};">
            <div class="res-card-header">
                <span class="res-id-badge" style="background-color: ${isActive ? '#27ae60' : '#c0392b'};">#${room.id}</span>
                <span style="font-weight:bold; color:#7f8c8d;">${room.number}</span>
            </div>
            <div class="res-room-info">${room.name}</div>
            <div style="font-size:0.85rem; color:#666; margin-bottom:8px;">${room.description || 'Sin descripción'}</div>
            
            <div class="res-dates">
                <div>🛏️ Cap: ${room.maxCapacity} | 🏷️ ${room.category}</div>
            </div>
            
            <div class="res-price">$${room.hourlyPrice} <span style="font-size:0.7rem; color:#aaa;">/hr</span></div>
            
            <button class="action-btn" style="margin-top:10px; width:100%;" onclick="prefillReservation(${room.id})">Reservar</button>
        </div>
        `;
    });
    html += '</div>';

    // Pagination Controls
    if (totalPages > 1) {
        html += `
        <div class="pagination">
             <button onclick="changeRoomsPage(-1)" ${currentRoomsPage === 1 ? 'disabled' : ''}>Prev</button>
            <span>Page ${currentRoomsPage} of ${totalPages} (Total: ${allRoomsData.length})</span>
            <button onclick="changeRoomsPage(1)" ${currentRoomsPage === totalPages ? 'disabled' : ''}>Next</button>
        </div>`;
    }

    container.innerHTML = html;
}

window.changeRoomsPage = function (delta) {
    currentRoomsPage += delta;
    renderRoomsPage();
}



window.prefillReservation = function (roomId) {
    showSection('reservations');
    document.getElementById('res-room-id').value = roomId;
    document.getElementById('res-room-id').scrollIntoView({ behavior: 'smooth', block: 'center' });
    const input = document.getElementById('res-room-id');
    input.style.borderColor = 'var(--accent)';
    setTimeout(() => input.style.borderColor = '', 2000);
}


// State for pagination
let currentReservationsPage = 1;
let allReservationsData = [];
const RESERVATIONS_PAGE_Size = 10;

async function searchReservations() {
    const from = document.getElementById('res-search-from').value;
    const to = document.getElementById('res-search-to').value;
    const guest = document.getElementById('res-search-guest').value;

    let params = new URLSearchParams();
    if (from) params.append('from', from + ":00.0-03:00");
    if (to) params.append('to', to + ":00.0-03:00");
    if (guest) params.append('guestId', guest);

    // We fetch everything (page 0, huge size to simulate "all")
    params.append('page', '0');
    params.append('pageSize', '2000');
    params.append('sortBy', 'STARTDATE');
    params.append('direction', 'ASC');

    document.getElementById('res-display-reservations').innerText = 'Loading all data...';

    try {
        const response = await fetch(`${RESERVATIONS_URL}/reservations?${params.toString()}`);
        if (!response.ok) throw new Error('Network response was not ok');

        const data = await response.json();
        allReservationsData = Array.isArray(data) ? data : (data.content || []);
        currentReservationsPage = 1;

        if (allReservationsData.length === 0) {
            document.getElementById('res-display-reservations').innerHTML = '<div class="p-4 bg-gray-100 rounded">No se encontraron reservas.</div>';
            return;
        }

        renderReservationsPage();
    } catch (error) {
        document.getElementById('res-display-reservations').innerHTML = `<div class="error-msg">Error: ${error.message}</div>`;
    }
}

function renderReservationsPage() {
    const container = document.getElementById('res-display-reservations');
    const start = (currentReservationsPage - 1) * RESERVATIONS_PAGE_Size;
    const end = start + RESERVATIONS_PAGE_Size;
    const pageItems = allReservationsData.slice(start, end);
    const totalPages = Math.ceil(allReservationsData.length / RESERVATIONS_PAGE_Size);

    let html = '<div class="reservations-grid">';
    pageItems.forEach(r => {
        const startDate = new Date(r.startDate).toLocaleString();
        const endDate = new Date(r.endDate).toLocaleString();
        const roomNum = r.room ? r.room.number : 'N/A';
        const roomType = r.room ? r.room.category : '';

        html += `
        <div class="reservation-card">
            <div class="res-card-header">
                <span class="res-id-badge">#${r.id}</span>
            </div>
            <div class="res-room-info">Habitación ${roomNum} <span style="font-size:0.8em; font-weight:normal; color:#666;">(${roomType})</span></div>
            <span class="res-guest-name">👤 ${r.guestId}</span>
            
            <div class="res-dates">
                <div>📅 In: ${startDate}</div>
                <div>📅 Out: ${endDate}</div>
            </div>
            
            <div class="res-price">$${r.totalPrice ? r.totalPrice.toFixed(2) : '0.00'}</div>
            <div style="font-size:0.75rem; color:#aaa; margin-top:5px; text-align:right;">ID R: ${r.room ? r.room.id : '?'} | Folder: ${r.billingFolderId ? r.billingFolderId.substring(0, 8) + '...' : 'N/A'}</div>
        </div>`;
    });
    html += '</div>';

    // Pagination Controls
    if (totalPages > 1) {
        html += `
        <div class="pagination">
            <button onclick="changeReservationsPage(-1)" ${currentReservationsPage === 1 ? 'disabled' : ''}>Prev</button>
            <span>Page ${currentReservationsPage} of ${totalPages} (Total: ${allReservationsData.length})</span>
            <button onclick="changeReservationsPage(1)" ${currentReservationsPage === totalPages ? 'disabled' : ''}>Next</button>
        </div>`;
    }

    container.innerHTML = html;
}

window.changeReservationsPage = function (delta) {
    currentReservationsPage += delta;
    renderReservationsPage();
}

async function createRoom() {
    const number = document.getElementById('room-create-number').value;
    const name = document.getElementById('room-create-name').value;
    const cat = document.getElementById('room-create-category').value;
    const price = document.getElementById('room-create-price').value;
    const cap = document.getElementById('room-create-capacity').value;
    const desc = document.getElementById('room-create-desc').value;

    document.getElementById('res-room-create').innerText = 'Sending...';

    const response = await fetch(`${RESERVATIONS_URL}/rooms`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            number: number,
            active: true,
            name: name,
            description: desc,
            maxCapacity: parseInt(cap),
            category: cat,
            hourlyPrice: parseFloat(price)
        })
    });
    renderResponse(response, 'res-room-create');
}

async function getRoomById() {
    const id = document.getElementById('room-search-id').value;
    document.getElementById('res-room-detail').innerText = 'Loading...';

    // Add some default availability params as per Bruno context to avoid errors if required
    const params = new URLSearchParams();
    const now = new Date();
    const nextYear = new Date();
    nextYear.setFullYear(now.getFullYear() + 1);

    // Formatting to ISO-like structure expected by backend if needed, 
    // but based on Bruno example: 2025-01-01T00:00:00-03:00
    // We'll use a simplified approach or just send the ID if params are optional.
    // Bruno says params are enabled=true.

    // Let's manually construct ISO string with timezone offset if possible, 
    // or just use hardcoded broad range for "Get Room Info" purposes.
    params.append('availabilityFrom', now.toISOString().split('.')[0] + "-03:00");
    params.append('availabilityTo', nextYear.toISOString().split('.')[0] + "-03:00");

    const response = await fetch(`${RESERVATIONS_URL}/rooms/${id}?${params.toString()}`);
    renderResponse(response, 'res-room-detail');
}
