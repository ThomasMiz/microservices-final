
// Room Service Microservice Logic

async function createOrder() {
    const roomId = document.getElementById('rs-room-id').value;
    const items = document.getElementById('rs-items').value.split(',');
    document.getElementById('res-rs-order').innerText = 'Sending request...';

    const response = await fetch(`${ROOM_SERVICE_URL}/orders`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            roomId: roomId,
            items: items.map(i => i.trim())
        })
    });

    renderResponse(response, 'res-rs-order');
}
