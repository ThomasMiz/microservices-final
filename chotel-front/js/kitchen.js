
// Kitchen Microservice Logic

async function getKitchenMenu() {
    const display = document.getElementById('kitchen-menu-display');
    display.innerHTML = '<div class="loading-spinner">Cargando inventario...</div>';

    try {
        const response = await fetch(`${config.KITCHEN_URL}/inventory`, {
            method: 'GET'
        });

        if (response.ok) {
            const items = await response.json();

            if (!items || items.length === 0) {
                display.innerHTML = '<div class="p-4 bg-gray-100 rounded">No hay ítems en el inventario.</div>';
                return;
            }

            let html = `
            <table class="inventory-table">
                <thead>
                    <tr>
                        <th>Item ID</th>
                        <th>Stock Actual</th>
                        <th>Última Actualización</th>
                    </tr>
                </thead>
                <tbody>`;

            items.forEach(item => {
                const stock = item.currentStock;
                const isLow = stock <= 10;
                const badgeClass = isLow ? 'stock-low' : '';

                html += `<tr>
                    <td><strong>${item.menuItemId}</strong></td>
                    <td><span class="stock-badge ${badgeClass}">${stock} unidades</span></td>
                    <td>${new Date(item.lastUpdated).toLocaleString()}</td>
                </tr>`;
            });
            html += '</tbody></table>';
            display.innerHTML = html;
        } else {
            renderResponse(response, 'kitchen-menu-display');
        }
    } catch (error) {
        display.innerHTML = `<span class="error-msg">Error de conexión: ${error.message}</span>`;
    }
}

async function setInventoryStock() {
    const itemId = document.getElementById('inv-item-id').value;
    const quantity = document.getElementById('inv-quantity').value;

    const data = {
        menuItemId: itemId,
        quantity: parseInt(quantity)
    };

    const response = await fetch(`${config.KITCHEN_URL}/inventory`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    renderResponse(response, 'res-inventory-update');
    getKitchenMenu(); // Refrescamos la tabla automáticamente
}

async function addInventoryStock() {
    const itemId = document.getElementById('inv-item-id-add').value;
    const quantityToAdd = document.getElementById('inv-quantity-add').value;

    const data = {
        menuItemId: itemId,
        quantity: parseInt(quantityToAdd)
    };

    const response = await fetch(`${config.KITCHEN_URL}/inventory/${itemId}/add`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    renderResponse(response, 'res-inventory-add');
    getKitchenMenu(); // Refrescamos la tabla
}
