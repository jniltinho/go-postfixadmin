<template>
  <div>
    <slot name="toolbar" />

    <div class="table-card">
      <div v-if="title || hasTitleAction" class="table-topbar table-titlebar">
        <div class="table-topbar-title">{{ title }}</div>
        <slot name="title-action" />
      </div>

      <div class="table-topbar table-controls">
        <div class="controls-left">
          <div class="per-page-wrap">
            <select v-model="rowsPerPage" class="ctrl-select" @change="currentPage = 1">
              <option :value="10">10</option>
              <option :value="15">15</option>
              <option :value="25">25</option>
              <option :value="50">50</option>
            </select>
            <span class="ctrl-label">entries per page</span>
          </div>
        </div>
        <div class="controls-right">
          <slot name="topbar-extra" />
          <span class="ctrl-label">Search:</span>
          <input v-model="search" class="search-input" placeholder="Search records..." @input="currentPage = 1" />
        </div>
      </div>

      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr class="table-head-row">
              <th v-for="col in columns" :key="col.key" class="table-th" @click="sortBy(col.key)">
                {{ col.label }}
                <span class="sort-arrows">
                  <span :class="{ 'sort-active': sortKey === col.key && sortDir === 'asc' }">▲</span>
                  <span :class="{ 'sort-active': sortKey === col.key && sortDir === 'desc' }">▼</span>
                </span>
              </th>
              <th v-if="showActions" class="table-th" style="text-align:right">ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td :colspan="columns.length + (showActions ? 1 : 0)" class="table-loading">
                <div class="spinner mx-auto" />
              </td>
            </tr>
            <tr v-else-if="pagedRows.length === 0">
              <td :colspan="columns.length + (showActions ? 1 : 0)" class="table-empty">No records found</td>
            </tr>
            <template v-else>
              <tr v-for="row in pagedRows" :key="(row as any)[keyField]" class="table-row">
                <td v-for="col in columns" :key="col.key" class="table-td">
                  <slot :name="`cell-${col.key}`" :row="row" :value="(row as any)[col.key]">
                    {{ (row as any)[col.key] ?? '—' }}
                  </slot>
                </td>
                <td v-if="showActions" class="table-td actions-td">
                  <slot name="actions" :row="row">
                    <button class="act-btn act-edit" @click="emit('edit', row)">
                      <Icon name="pencil" :size="12" style="margin-right:4px;vertical-align:middle" />EDIT
                    </button>
                    <button class="act-btn act-del" @click="emit('delete', row)">
                      <Icon name="trash-2" :size="12" style="margin-right:4px;vertical-align:middle" />DELETE
                    </button>
                  </slot>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <div class="table-footer">
        <div class="showing-text">
          <template v-if="filteredRows.length === 0">Showing 0 entries</template>
          <template v-else>
            Showing {{ (currentPage - 1) * rowsPerPage + 1 }} to
            {{ Math.min(currentPage * rowsPerPage, filteredRows.length) }} of
            {{ filteredRows.length }} entries
          </template>
        </div>
        <div class="pagination">
          <button class="pg-btn" :disabled="currentPage === 1" @click="currentPage = 1">FIRST</button>
          <button class="pg-btn" :disabled="currentPage === 1" @click="currentPage--">PREVIOUS</button>
          <button
            v-for="p in pageButtons" :key="p"
            class="pg-btn" :class="{ 'pg-active': p === currentPage }"
            @click="currentPage = p"
          >{{ p }}</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage++">NEXT</button>
          <button class="pg-btn" :disabled="currentPage === totalPages" @click="currentPage = totalPages">LAST</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, useSlots } from 'vue'

interface Column {
  key: string
  label: string
}

const slots = useSlots()

const props = withDefaults(defineProps<{
  rows: any[]
  columns: Column[]
  rowKey?: string
  searchFields?: string[]
  defaultSortKey?: string
  defaultSortDir?: 'asc' | 'desc'
  loading?: boolean
  showActions?: boolean
  title?: string
  initialRowsPerPage?: number
}>(), {
  showActions: true,
  initialRowsPerPage: 15,
})

const emit = defineEmits<{
  edit: [row: any]
  delete: [row: any]
}>()

const keyField = computed(() => props.rowKey ?? props.columns[0]?.key ?? 'id')
const hasTitleAction = computed(() => Boolean(slots['title-action']))

const search = ref('')
const rowsPerPage = ref(props.initialRowsPerPage)
const currentPage = ref(1)
const sortKey = ref(props.defaultSortKey ?? props.columns[0]?.key ?? '')
const sortDir = ref<'asc' | 'desc'>(props.defaultSortDir ?? 'asc')

function sortBy(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

const filteredRows = computed(() => {
  let rows = props.rows
  const q = search.value.toLowerCase()
  if (q) {
    const fields = props.searchFields ?? props.columns.map(c => c.key)
    rows = rows.filter(r => fields.some(f => String((r as any)[f] ?? '').toLowerCase().includes(q)))
  }
  return [...rows].sort((a, b) => {
    const av = String((a as any)[sortKey.value] ?? '')
    const bv = String((b as any)[sortKey.value] ?? '')
    return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredRows.value.length / rowsPerPage.value)))

const pageButtons = computed(() => {
  const total = totalPages.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const cur = currentPage.value
  const pages = new Set([1, total, cur, cur - 1, cur + 1].filter(p => p >= 1 && p <= total))
  return Array.from(pages).sort((a, b) => a - b)
})

const pagedRows = computed(() => {
  const start = (currentPage.value - 1) * rowsPerPage.value
  return filteredRows.value.slice(start, start + rowsPerPage.value)
})

watch([search, rowsPerPage, () => props.rows], () => { currentPage.value = 1 })
</script>

<style scoped>

.table-card {
  background: #fff;
  border: 2px solid #1e293b;
  box-shadow: 2px 2px 0 #1e293b;
  overflow: hidden;
  padding: 16px 8px 4px;
}

.table-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0 12px;
  border-bottom: none;
  background: transparent;
  font-size: 12px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.table-topbar-title {
  font-weight: 900;
  color: #1e293b;
}

.table-titlebar + .table-controls {
  border-top: 1px solid #e2e8f0;
  padding-top: 12px;
  margin-top: 8px;
}

.table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0 0;
  border-top: none;
  background: transparent;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
  border: 2px solid #1e293b;
}

.table-head-row {
  background: #3b82f6;
  border-bottom: 2px solid #1e293b;
}

.table-th {
  color: #fff;
  font-weight: 700;
  font-size: 12px;
  letter-spacing: 0.4px;
  padding: 10px;
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
  font-family: var(--font-sans);
}

.table-th:hover { background: #2563eb; }

tbody tr {
  border-bottom: 1px solid #e2e8f0;
}

tbody tr:last-child {
  border-bottom: none;
}

tbody tr:nth-child(even) {
  background: #f8fafc;
}

tbody tr:hover {
  background: #eff6ff;
}

.table-td {
  padding: 7px 10px;
  color: #1e293b;
  font-size: 12px;
}

.actions-td {
  display: flex;
  gap: 6px;
  align-items: center;
  justify-content: flex-end;
}

.td-link {
  color: #1e293b;
  font-weight: 600;
}

.table-loading,
.table-empty {
  text-align: center;
  padding: 28px;
  color: #94a3b8;
  font-size: 13px;
}

.controls-left,
.controls-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.per-page-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
}

.ctrl-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.ctrl-select,
.search-input {
  border: 2px solid #1e293b !important;
  padding: 4px 8px !important;
  font-size: 13px !important;
  color: #1e293b !important;
  background-color: #ffffff !important;
  outline: none !important;
  border-radius: 0 !important;
  box-sizing: border-box !important;
  font-weight: 600;
  transition: border-color 0.15s ease-in-out;
}

.ctrl-select:focus,
.search-input:focus {
  border-color: #3b82f6 !important;
}

.search-input {
  width: 200px;
}

.showing-text {
  font-size: 12.5px;
  color: #64748b;
  font-weight: 500;
}

.pagination {
  display: flex;
  gap: 3px;
  align-items: center;
}

.pg-btn {
  height: 28px;
  padding: 0 10px;
  font-size: 10px;
  font-weight: 700;
  color: #374151;
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 0;
  cursor: pointer;
  letter-spacing: 0.4px;
  text-transform: uppercase;
  white-space: nowrap;
  transition: all 0.1s ease-in-out;
}

.pg-btn:hover:not(:disabled) {
  border-color: #1e293b;
  color: #1e293b;
  background: #f8fafc;
}

.pg-btn:disabled {
  opacity: 0.35;
  cursor: default;
}

.pg-active {
  background: #3b82f6 !important;
  color: #ffffff !important;
  border-color: #3b82f6 !important;
}

.act-btn,
:slotted(.act-btn) {
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
  border: 1px solid #1e293b;
  letter-spacing: 0.4px;
  border-radius: 0;
  display: inline-flex;
  align-items: center;
  transition: all .12s;
  box-shadow: 1px 1px 0 #1e293b;
  text-transform: uppercase;
}

.act-btn:hover,
:slotted(.act-btn:hover) { transform: translate(-0.5px, -0.5px); }
.act-btn:active,
:slotted(.act-btn:active) { transform: translate(0,0); box-shadow: none; }
.act-edit,
:slotted(.act-edit) { background: oklch(0.546 0.245 262.881); color: #fff; }
.act-edit:hover,
:slotted(.act-edit:hover) { background: #fff; color: oklch(0.546 0.245 262.881); }
.act-del,
:slotted(.act-del) { background: oklch(0.577 0.245 27.325); color: #fff; }
.act-del:hover,
:slotted(.act-del:hover) { background: #fff; color: oklch(0.577 0.245 27.325); }


:slotted(.badge-yes) {
  background: oklch(0.962 0.044 156.743);
  color: oklch(0.527 0.154 150.069);
  border: 2px solid oklch(0.527 0.154 150.069);
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 700;
  display: inline-block;
}

:slotted(.badge-no) {
  background: #fee2e2;
  color: #dc2626;
  border: 2px solid #dc2626;
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 700;
  display: inline-block;
}

:slotted(.act-disabled) { opacity: .4; cursor: not-allowed; pointer-events: none; }

:slotted(.cell-with-icon) {
  display: flex;
  align-items: center;
  gap: 6px;
}

:slotted(.row-icon) {
  color: #3b82f6;
  flex-shrink: 0;
}

</style>
