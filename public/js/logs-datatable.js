$(document).ready(function () {
    var dtLang = {};
    var currentLang = window.AppLang || "en";
    if (currentLang === "pt" || currentLang === "pt_BR") {
        dtLang = { url: "/static/js/i18n/pt-BR.json" };
    } else if (currentLang === "es") {
        dtLang = { url: "/static/js/i18n/es-ES.json" };
    }

    if ($('#logsTable').length) {
        var table = $('#logsTable').DataTable({
            "processing": true,
            "serverSide": true,
            "ajax": {
                "url": "/api/logs",
                "type": "GET"
            },
            "columns": [
                { "data": "timestamp", "className": "text-gray-600 font-medium text-xs py-1 px-2" },
                { "data": "username", "className": "text-brand-primary font-bold text-xs py-1 px-2" },
                { "data": "domain", "className": "text-gray-600 font-medium text-xs py-1 px-2" },
                { "data": "action", "className": "uppercase text-xs font-black tracking-wide py-1 px-2" },
                { "data": "data", "className": "text-gray-600 font-mono text-xs py-1 px-2" }
            ],
            "order": [[0, "desc"]],
            "language": dtLang,
            "pageLength": 10,
            "drawCallback": function (settings) {
                // Re-initialize Lucide icons when data redraws
                if (typeof lucide !== 'undefined') {
                    lucide.createIcons();
                }
            }
        });
    }

    var standardTables = ['#mailboxesTable', '#domainsTable', '#adminsTable', '#aliasesTable', '#aliasDomainsTable'];
    standardTables.forEach(function (selector) {
        if ($(selector).length) {
            $(selector).DataTable({
                language: dtLang,
                pageLength: 10,
                // Disable ordering on the last column (actions)
                columnDefs: [
                    { orderable: false, targets: -1 }
                ],
                drawCallback: function (settings) {
                    // Re-render Lucide icons on pagination/draw
                    if (window.lucide) {
                        lucide.createIcons();
                    }
                }
            });
        }
    });

});
