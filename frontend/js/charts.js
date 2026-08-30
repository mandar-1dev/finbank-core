// charts.js — thin wrappers around Chart.js so page scripts stay short.
// Chart.js is loaded from cdnjs in each banker page that needs it.

function renderLineChart(canvasId, labels, data, label) {
  const ctx = document.getElementById(canvasId);
  if (!ctx || typeof Chart === 'undefined') return null;
  return new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label,
        data,
        borderColor: '#146c56',
        backgroundColor: 'rgba(20,108,86,0.12)',
        tension: 0.35,
        fill: true,
        pointRadius: 3,
      }],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { display: false } },
      scales: { y: { beginAtZero: true } },
    },
  });
}

function renderDoughnutChart(canvasId, labels, data, colors) {
  const ctx = document.getElementById(canvasId);
  if (!ctx || typeof Chart === 'undefined') return null;
  return new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels,
      datasets: [{ data, backgroundColor: colors || ['#146c56', '#d9a441', '#c0392b', '#2266b0'] }],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { position: 'bottom' } },
    },
  });
}

function renderBarChart(canvasId, labels, data, label) {
  const ctx = document.getElementById(canvasId);
  if (!ctx || typeof Chart === 'undefined') return null;
  return new Chart(ctx, {
    type: 'bar',
    data: { labels, datasets: [{ label, data, backgroundColor: '#146c56', borderRadius: 4 }] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { display: false } },
      scales: { y: { beginAtZero: true } },
    },
  });
}
