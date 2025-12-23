window.APP_CONFIG = window.APP_CONFIG || {
    BASE_URL: DEFAULT_API_BASE,
    LOBBY_URL: `${DEFAULT_API_BASE}/lobby`,
    RESERVATIONS_URL: `${DEFAULT_API_BASE}/reservations`,
    CLEANING_URL: `${DEFAULT_API_BASE}/cleaning`,
    BILLING_URL: `${DEFAULT_API_BASE}/billing`,
    ROOM_SERVICE_URL: `${DEFAULT_API_BASE}/roomservice`,
    KITCHEN_URL: `${DEFAULT_API_BASE}/kitchen`
};

const config = window.APP_CONFIG;

window.LOBBY_URL = config.LOBBY_URL;
window.RESERVATIONS_URL = config.RESERVATIONS_URL;
window.CLEANING_URL = config.CLEANING_URL;
window.BILLING_URL = config.BILLING_URL;
window.ROOM_SERVICE_URL = config.ROOM_SERVICE_URL;
window.KITCHEN_URL = config.KITCHEN_URL;

window.CLEANING_AUTH = "Basic TWFyY2VsbzpjaG90YXJkbzY5";

function showSection(id) {
    document.querySelectorAll('.section').forEach(s => s.classList.remove('active'));
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
    document.getElementById(id).classList.add('active');
    event.currentTarget.classList.add('active');
}

async function renderResponse(response, displayId) {
    const container = document.getElementById(displayId);
    const status = response.status;
    let data;

    try {
        data = await response.json();
    } catch (e) {
        data = { message: "Respuesta sin cuerpo JSON o error al parsear" };
    }

    container.innerHTML = `<strong>Status: ${status}</strong>\n${JSON.stringify(data, null, 2)}`;
}
