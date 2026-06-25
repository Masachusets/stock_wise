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
