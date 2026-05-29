<template>
  <DataTable
    :data="data"
    :columns="processedColumns"
    :options="finalOptions"
    class="brutal-datatable"
    @init="onInit"
  />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
// Using loose typing for rapid integration with datatables.net-vue3
type DataTableOptions = any

interface ColumnDef {
  data?: string
  name?: string
  title?: string
  className?: string
  orderable?: boolean
  render?: (data: any, type: string, row: any) => string
}

const props = defineProps<{
  data: any[]
  columns: ColumnDef[]
  options?: Partial<DataTableOptions>
  pageLength?: number
  language?: 'PT' | 'EN' | 'ES'
}>()

const emit = defineEmits<{
  (e: 'row-click', row: any): void
  (e: 'draw'): void
}>()

const langMap: Record<string, any> = {
  PT: {
    emptyTable: "Nenhum registro encontrado",
    info: "Mostrando de _START_ até _END_ de _TOTAL_ registros",
    infoFiltered: "(Filtrados de _MAX_ registros)",
    loadingRecords: "Carregando...",
    zeroRecords: "Nenhum registro encontrado",
    search: "Pesquisar:",
    lengthMenu: "Mostrar _MENU_ registros",
    paginate: {
      next: "Próximo",
      previous: "Anterior",
      first: "Primeiro",
      last: "Último"
    }
  },
  EN: {
    search: "Search:",
    lengthMenu: "Show _MENU_ entries"
  },
  ES: {
    search: "Buscar:",
    lengthMenu: "Mostrar _MENU_ registros"
  }
}

const finalOptions = computed(() => {
  const base: any = {
    pageLength: props.pageLength || 15,
    lengthMenu: [10, 15, 25, 50],
    order: [[0, 'asc']],
    columnDefs: [
      { orderable: false, targets: -1 } // last column (actions) not sortable
    ],
    drawCallback: (_settings: any) => {
      // Re-init Lucide icons after every draw (pagination, search, sort)
      if (typeof window !== 'undefined' && (window as any).lucide) {
        (window as any).lucide.createIcons()
      }
      emit('draw')
    },
    language: langMap[props.language || 'EN'] || langMap.EN,
    ...props.options
  }
  return base
})

const processedColumns = computed(() => {
  return props.columns.map(col => ({
    ...col,
    // If render is provided as function, DataTables accepts it
  }))
})

function onInit(_dt: any) {
  // dt is the DataTables API instance — useful for advanced control later
  // console.log('DataTable initialized', dt)
}

onMounted(() => {
  // nothing extra for now
})
</script>

<style scoped>
/* Make sure the DataTables container respects our .table-card */
:deep(.dt-container) {
  width: 100%;
}

:deep(table.dataTable) {
  border-collapse: collapse;
}

:deep(table.dataTable thead) {
  background: #3b82f6;
  color: white;
}

:deep(table.dataTable thead th) {
  font-weight: 700;
  font-size: 12px;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  padding: 10px !important;
  border-bottom: 3px solid #1e293b;
}

:deep(table.dataTable tbody td) {
  padding: 8px 10px;
  font-size: 13px;
  border-bottom: 1px solid #e5e7eb;
}

:deep(.dt-layout-row) {
  margin: 8px 0;
}

:deep(.dt-length, .dt-search) {
  font-size: 12px;
  font-weight: 600;
}

:deep(.dt-search input) {
  border: 2px solid #1e293b;
  padding: 4px 8px;
  font-size: 13px;
}

:deep(.dt-paging .dt-paging-button) {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  border: 2px solid transparent;
  padding: 4px 10px;
}

:deep(.dt-paging .dt-paging-button.current) {
  background: #3b82f6;
  color: white;
  border-color: #1e293b;
}
</style>