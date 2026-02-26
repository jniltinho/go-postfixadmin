$(document).ready(function () {
    var dtLang = {};
    var currentLang = window.AppLang || "en";
    if (currentLang === "pt" || currentLang === "pt_BR") {
        dtLang = {
            "sEmptyTable": "Nenhum registro encontrado",
            "sInfo": "Mostrando de _START_ até _END_ de _TOTAL_ registros",
            "sInfoEmpty": "Mostrando 0 até 0 de 0 registros",
            "sInfoFiltered": "(Filtrados de _MAX_ registros)",
            "sInfoPostFix": "",
            "sInfoThousands": ".",
            "sLengthMenu": "_MENU_ resultados por página",
            "sLoadingRecords": "Carregando...",
            "sProcessing": "Processando...",
            "sZeroRecords": "Nenhum registro encontrado",
            "sSearch": "Pesquisar:",
            "oPaginate": {
                "sNext": "Próximo",
                "sPrevious": "Anterior",
                "sFirst": "Primeiro",
                "sLast": "Último"
            },
            "oAria": {
                "sSortAscending": ": Ordenar colunas de forma ascendente",
                "sSortDescending": ": Ordenar colunas de forma descendente"
            }
        };
    } else if (currentLang === "es") {
        dtLang = {
            "sProcessing": "Procesando...",
            "sLengthMenu": "Mostrar _MENU_ registros",
            "sZeroRecords": "No se encontraron resultados",
            "sEmptyTable": "Ningún dato disponible en esta tabla",
            "sInfo": "Mostrando registros del _START_ al _END_ de un total de _TOTAL_ registros",
            "sInfoEmpty": "Mostrando registros del 0 al 0 de un total de 0 registros",
            "sInfoFiltered": "(filtrado de un total de _MAX_ registros)",
            "sSearch": "Buscar:",
            "oPaginate": {
                "sFirst": "Primero",
                "sLast": "Último",
                "sNext": "Siguiente",
                "sPrevious": "Anterior"
            },
        };
    }

    var table = $('#logsTable').DataTable({
        "processing": true,
        "serverSide": true,
        "ajax": {
            "url": "/api/logs",
            "type": "GET"
        },
        "columns": [
            { "data": "timestamp", "className": "text-gray-600 font-medium text-xs py-2 px-4" },
            { "data": "username", "className": "text-brand-primary font-bold text-xs py-2 px-4" },
            { "data": "domain", "className": "text-gray-600 font-medium text-xs py-2 px-4" },
            { "data": "action", "className": "uppercase text-xs font-black tracking-wide py-2 px-4" },
            { "data": "data", "className": "text-gray-600 font-mono text-xs py-2 px-4" }
        ],
        "order": [[0, "desc"]],
        "language": dtLang,
        "pageLength": 12,
        "drawCallback": function (settings) {
            // Re-initialize Lucide icons when data redraws
            if (typeof lucide !== 'undefined') {
                lucide.createIcons();
            }
        }
    });

});
