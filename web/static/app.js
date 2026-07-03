function filterTable() {
    const search = document.getElementById('search').value.toLowerCase();
    const status = document.getElementById('statusFilter').value;
    const rows = document.querySelectorAll('#equipmentTable tbody tr');

    rows.forEach(row => {
        const text = row.textContent.toLowerCase();
        const rowStatus = row.querySelector('.badge')?.textContent.toLowerCase() || '';
        const matchSearch = !search || text.includes(search);
        const matchStatus = !status || row.classList.contains('badge-' + status) || row.querySelector('.badge')?.classList.contains('badge-' + status);
        row.style.display = (matchSearch && matchStatus) ? '' : 'none';
    });
}

function signWaybill(id) {
    if (!confirm('Подписать накладную?')) return;
    fetch('/api/waybills/' + id + '/sign', { method: 'POST' })
        .then(r => r.ok ? location.reload() : alert('Ошибка'));
}

function archiveWaybill(id) {
    if (!confirm('Архивировать накладную?')) return;
    fetch('/api/waybills/' + id + '/archive', { method: 'POST' })
        .then(r => r.ok ? location.reload() : alert('Ошибка'));
}

function deleteWaybill(id) {
    if (!confirm('Удалить накладную?')) return;
    fetch('/api/waybills/' + id, { method: 'DELETE' })
        .then(r => r.ok ? location.reload() : alert('Ошибка'));
}

function showEditModal() {
    document.getElementById('editModal').style.display = 'flex';
}

function hideEditModal() {
    document.getElementById('editModal').style.display = 'none';
}

function deleteEquipment(invNum) {
    if (!confirm('Удалить оборудование ' + invNum + '?')) return;
    fetch('/equipments/' + invNum, { method: 'DELETE' })
        .then(r => {
            if (r.ok) window.location = '/equipments';
            else r.text().then(t => alert('Ошибка: ' + t));
        });
}

function submitEdit(e) {
    e.preventDefault();
    const form = e.target;
    const fd = new FormData(form);
    const data = {};
    fd.forEach((v, k) => {
        if (!v) return;
        if (k === 'nomenclature_id') data[k] = parseInt(v, 10);
        else data[k] = v;
    });
    const invNum = form.querySelector('[name="inventory_number"]').value;

    fetch('/equipments/' + invNum + '/update', {
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data)
    })
    .then(r => {
        if (r.ok) location.reload();
        else r.text().then(t => alert('Ошибка: ' + t));
    });
}

function showModal() {
    document.getElementById('modal').style.display = 'flex';
}

function hideModal() {
    document.getElementById('modal').style.display = 'none';
}

function submitForm(e) {
    e.preventDefault();
    const form = e.target;

    // Validate dates
    const mfg = document.getElementById('manufacture_date').value;
    const arr = document.getElementById('arrival_date').value;
    if (mfg && arr && mfg > arr) {
        alert('Дата изготовления должна быть раньше даты поступления');
        return;
    }

    const data = {};
    const fd = new FormData(form);
    fd.forEach((v, k) => {
        if (!v) return;
        // Преобразуем nomenclature_id в число
        if (k === 'nomenclature_id') {
            data[k] = parseInt(v, 10);
        } else {
            data[k] = v;
        }
    });

    // Добавляем "ИТ" к инвентарному номеру
    const invInput = form.querySelector('[name="inventory_number"]');
    data['inventory_number'] = 'ИТ' + invInput.value;

    fetch('/equipments/add', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data)
    })
    .then(r => {
        if (r.ok) location.reload();
        else r.text().then(t => alert('Ошибка: ' + t));
    });
}
