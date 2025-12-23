// Lobby Microservice Logic

async function doCheckIn() {
    const room = document.getElementById('lobby-checkin-room').value;
    const guestId = document.getElementById('lobby-checkin-guest').value;

    const response = await fetch(`${LOBBY_URL}/check_in/${room}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guest_id: guestId })
    });

    renderResponse(response, 'res-lobby-checkin');
}

async function doCheckOut() {
    const room = document.getElementById('lobby-checkout-room').value;

    const response = await fetch(`${LOBBY_URL}/check_out/${room}`, {
        method: 'POST'
    });

    renderResponse(response, 'res-lobby-checkout');
}
