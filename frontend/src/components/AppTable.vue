<template>
  <div>
    <slot name="toolbar" />

    <div class="table-card">
      <div class="table-topbar">
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
import { ref, computed, watch } from 'vue'

interface Column {
  key: string
  label: string
}

const props = withDefaults(defineProps<{
  rows: any[]
  columns: Column[]
  rowKey?: string
  searchFields?: string[]
  defaultSortKey?: string
  defaultSortDir?: 'asc' | 'desc'
  loading?: boolean
  showActions?: boolean
}>(), {
  showActions: true,
})

const emit = defineEmits<{
  edit: [row: any]
  delete: [row: any]
}>()

const keyField = computed(() => props.rowKey ?? props.columns[0]?.key ?? 'id')

const search = ref('')
const rowsPerPage = ref(15)
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
