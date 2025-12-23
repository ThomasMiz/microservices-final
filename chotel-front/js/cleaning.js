
// Cleaning Microservice Logic

async function assignCleaningJob() {
    const jobId = document.getElementById('clean-job-id').value;
    const staffId = document.getElementById('clean-staff-id').value;
    document.getElementById('res-clean-assign').innerText = 'Sending request...';

    const response = await fetch(`${CLEANING_URL}/jobs/${jobId}/assign`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': CLEANING_AUTH
        },
        body: JSON.stringify({ staffId: parseInt(staffId) })
    });

    renderResponse(response, 'res-clean-assign');
}

async function reportDamage() {
    const room = document.getElementById('clean-damage-room').value;
    const item = document.getElementById('clean-damage-item').value;
    const desc = document.getElementById('clean-damage-desc').value;
    document.getElementById('res-clean-damage').innerText = 'Sending request...';

    const response = await fetch(`${CLEANING_URL}/damage-reports`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': CLEANING_AUTH
        },
        body: JSON.stringify({
            roomNumber: parseInt(room),
            brokenItem: item,
            description: desc,
            fineAmount: 0.0 // Valor por defecto
        })
    });

    renderResponse(response, 'res-clean-damage');
}

async function createStaff() {
    const name = document.getElementById('clean-staff-name').value;
    document.getElementById('res-clean-staff-create').innerText = 'Sending...';

    const response = await fetch(`${CLEANING_URL}/staff`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': CLEANING_AUTH
        },
        body: JSON.stringify({ name: name })
    });
    renderResponse(response, 'res-clean-staff-create');
}

async function searchStaff() {
    document.getElementById('res-clean-staff-list').innerText = 'Loading...';
    const response = await fetch(`${CLEANING_URL}/staff?page=0&pageSize=50`, {
        headers: { 'Authorization': CLEANING_AUTH }
    });

    if (response.ok) {
        const data = await response.json(); // May be array or paged object
        const list = Array.isArray(data) ? data : (data.content || []);
        let html = '<ul>';
        list.forEach(s => html += `<li>ID: ${s.staffId} - ${s.name}</li>`);
        html += '</ul>';
        document.getElementById('res-clean-staff-list').innerHTML = html;
    } else {
        renderResponse(response, 'res-clean-staff-list');
    }
}

async function searchJobs() {
    const room = document.getElementById('clean-job-search-room').value;
    let url = `${CLEANING_URL}/jobs`;
    if (room) url += `?roomNumber=${room}`;

    const container = document.getElementById('res-clean-jobs-list');
    container.innerHTML = '<div class="loading-spinner">Cargando trabajos...</div>';

    try {
        const response = await fetch(url, {
            headers: { 'Authorization': CLEANING_AUTH }
        });

        if (response.ok) {
            const data = await response.json();
            // The API might return a Page object or List depending on implementation. 
            // Controller returns List<CleaningJobJson> directly with X-Total-Elements header.
            // Let's handle if it returns a list directly.
            const jobs = Array.isArray(data) ? data : (data.content || []);

            if (!jobs || jobs.length === 0) {
                container.innerHTML = '<div class="p-4 bg-gray-100 rounded">No search results found.</div>';
                return;
            }

            let html = `
            <table class="inventory-table">
                <thead>
                    <tr>
                        <th>Job ID</th>
                        <th>Habitación</th>
                        <th>Staff Asignado</th>
                        <th>Estado</th>
                        <th>Creado</th>
                    </tr>
                </thead>
                <tbody>`;

            jobs.forEach(job => {
                let status = 'PENDIENTE';
                let statusClass = 'state-pending'; // Reusing ticket state styles or similar

                if (job.finishedAt) {
                    status = 'TERMINADO';
                    statusClass = 'state-closed';
                } else if (job.startedAt) {
                    status = 'EN PROGRESO';
                    statusClass = 'res-status-active';
                }

                const staffName = job.assignedTo ? job.assignedTo.name : '<span style="color:#999; font-style:italic;">Sin asignar</span>';

                html += `<tr>
                    <td><strong>${job.id}</strong></td>
                    <td>${job.roomNumber}</td>
                    <td>${staffName}</td>
                    <td><span class="state-badge ${statusClass}">${status}</span></td>
                    <td>${new Date(job.createdAt).toLocaleString()}</td>
                </tr>`;
            });
            html += '</tbody></table>';
            container.innerHTML = html;
        } else {
            renderResponse(response, 'res-clean-jobs-list');
        }
    } catch (e) {
        container.innerHTML = `<span class="error-msg">Error: ${e.message}</span>`;
    }
}

async function markJobStarted() {
    const jobId = document.getElementById('clean-update-job-id').value;
    const staffId = document.getElementById('clean-update-staff-id').value;
    document.getElementById('res-clean-job-update').innerText = 'Sending...';

    const response = await fetch(`${CLEANING_URL}/jobs/${jobId}/started`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': CLEANING_AUTH
        },
        body: JSON.stringify({ staffId: parseInt(staffId) })
    });
    renderResponse(response, 'res-clean-job-update');
}

async function markJobFinished() {
    const jobId = document.getElementById('clean-update-job-id').value;
    const staffId = document.getElementById('clean-update-staff-id').value;
    document.getElementById('res-clean-job-update').innerText = 'Sending...';

    const response = await fetch(`${CLEANING_URL}/jobs/${jobId}/finished`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': CLEANING_AUTH
        },
        body: JSON.stringify({ staffId: parseInt(staffId) })
    });
    renderResponse(response, 'res-clean-job-update');
}

async function searchDamageReports() {
    document.getElementById('res-clean-damage-list').innerText = 'Loading...';
    const response = await fetch(`${CLEANING_URL}/damage-reports`, {
        headers: { 'Authorization': CLEANING_AUTH }
    });
    renderResponse(response, 'res-clean-damage-list');
}

async function testCleaningAuth() {
    document.getElementById('res-clean-auth').innerText = 'Testing...';
    const response = await fetch(`${CLEANING_URL}/auth/test`, {
        headers: { 'Authorization': CLEANING_AUTH }
    });
    renderResponse(response, 'res-clean-auth');
}

async function getJobById() {
    const id = document.getElementById('clean-job-search-id').value;
    document.getElementById('res-clean-jobs-list').innerText = 'Loading...';

    const response = await fetch(`${CLEANING_URL}/jobs/${id}`, {
        headers: { 'Authorization': CLEANING_AUTH }
    });
    renderResponse(response, 'res-clean-jobs-list');
}

async function getStaffById() {
    const id = document.getElementById('clean-search-staff-id').value;
    document.getElementById('res-clean-staff-list').innerText = 'Loading...';

    const response = await fetch(`${CLEANING_URL}/staff/${id}`, {
        headers: { 'Authorization': CLEANING_AUTH }
    });
    renderResponse(response, 'res-clean-staff-list');
}
