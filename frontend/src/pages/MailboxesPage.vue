<template>
  <div class="dom-page">

    <!-- ─── Header ─── -->
    <div class="dom-header">
      <div>
        <div class="dom-title">EMAIL ACCOUNTS</div>
        <div class="dom-subtitle">MANAGE YOUR EMAIL ACCOUNTS</div>
      </div>
      <button class="btn-primary" @click="openAdd">
        <Icon name="plus-circle" :size="16" style="margin-right:6px;vertical-align:middle" />
        ADD EMAIL ACCOUNT
      </button>
    </div>

    <!-- ─── Error banner ─── -->
    <div v-if="error" class="error-banner">
      <Icon name="alert-triangle" :size="16" class="mr-1" /> {{ error }}
    </div>

    <!-- Domain filter (kept like in old server, placed before table) -->
    <div class="mb-4 flex items-center gap-3">
      <label class="text-xs font-black uppercase tracking-widest text-brand-text">FILTER BY DOMAIN:</label>
      <select v-model="domainFilter" class="h-9 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium bg-white" @change="onFilterChange">
        <option value="">All Domains</option>
        <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
      </select>
      <a v-if="domainFilter" class="ml-2 text-xs font-bold text-red-500 hover:underline cursor-pointer" @click="domainFilter = ''; onFilterChange()">Clear Filter</a>
    </div>

    <!-- ─── Table card ─── -->
    <div class="table-card">

      <!-- Controls row -->
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
          <span class="ctrl-label">Search:</span>
          <input v-model="search" class="search-input" placeholder="Search records..." @input="currentPage = 1" />
        </div>
      </div>

      <!-- Table -->
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
              <th class="table-th">ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td :colspan="columns.length + 1" class="table-loading">
                <div class="spinner mx-auto" />
              </td>
            </tr>
            <tr v-else-if="pagedRows.length === 0">
              <td :colspan="columns.length + 1" class="table-empty">No records found</td>
            </tr>
            <tr v-for="row in pagedRows" :key="row.username" class="table-row">
              <td class="table-td td-link">
                <div class="cell-with-icon">
                  <Icon name="mail" :size="14" class="row-icon" />
                  {{ row.username }}
                </div>
              </td>
              <td class="table-td">{{ row.name || '—' }}</td>
              <td class="table-td mono">{{ row.domain }}</td>
              <td class="table-td">
                {{ row.quota === 0 ? 'Unlimited' : (row.quota / 1048576) + ' MB' }}
              </td>
              <td class="table-td">
                <span :class="row.active ? 'badge-yes' : 'badge-no'">{{ row.active ? 'YES' : 'NO' }}</span>
              </td>
              <td class="table-td">{{ formatDate(row.modified) }}</td>
              <td class="table-td actions-td">
                <button class="act-btn act-edit" @click="openEdit(row)">
                  <Icon name="pencil" :size="12" style="margin-right:4px;vertical-align:middle" />EDIT
                </button>
                <button class="act-btn act-del" @click="confirmDelete(row)">
                  <Icon name="trash-2" :size="12" style="margin-right:4px;vertical-align:middle" />DELETE
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Footer -->
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

    <!-- ══════════ ADD MODAL (from zero - exact structure from form_add_mailbox.html) ══════════ -->
    <div v-if="showAdd" class="modal-overlay" @click.self="closeAdd">
      <div class="bg-white border-4 border-brand-text w-full max-w-2xl max-h-[90vh] flex flex-col">
        <!-- Header (exact blue bar from reference) -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="plus-circle" :size="20" class="mr-2" />
            ADD EMAIL ACCOUNT
          </h3>
          <button @click="closeAdd" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="overflow-y-auto flex-1">
          <!-- Error Area (exact match) -->
          <div v-if="addError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
            <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-red-700 font-medium">{{ addError }}</p>
          </div>

          <form class="p-6 space-y-5" @submit.prevent="submitAdd">
            <!-- Email Account Section (border-2 exact) -->
            <div class="border-2 border-brand-text p-4 space-y-4">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="mail" :size="16" class="mr-2" />
                EMAIL ACCOUNT
              </h4>

              <!-- Username + Domain grid -->
              <div class="grid grid-cols-2 gap-3 items-start">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    USERNAME <span class="text-red-500">*</span>
                  </label>
                  <input
                    v-model="addForm.localPart"
                    type="text"
                    required
                    minlength="4"
                    pattern="[a-z0-9._-]{4,}"
                    placeholder="username"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm lowercase"
                    style="text-transform: lowercase;"
                    @input="onLocalPartInput"
                  />
                  <p class="text-[10px] text-gray-400 mt-1">Lowercase letters, numbers, dots, hyphens (min. 4 chars)</p>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">
                    DOMAIN <span class="text-red-500">*</span>
                  </label>
                  <select
                    v-model="addForm.domain"
                    required
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors cursor-pointer text-sm"
                    @change="updateEmailPreview"
                  >
                    <option value="" disabled>Select a domain...</option>
                    <option v-for="d in domains" :key="d.domain" :value="d.domain">{{ d.domain }}</option>
                  </select>
                </div>
              </div>

              <!-- Email Preview (exact styling from reference) -->
              <div class="bg-gray-50 border border-gray-300 px-3 h-10 flex items-center gap-2">
                <span class="text-xs font-bold uppercase tracking-widest text-gray-400 whitespace-nowrap">FULL ADDRESS:</span>
                <p class="font-mono text-sm font-bold">
                  <span class="text-brand-primary">{{ addForm.localPart || 'user' }}</span>@<span class="text-brand-primary">{{ addForm.domain || 'domain.com' }}</span>
                </p>
              </div>

              <!-- Display Name -->
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DISPLAY NAME</label>
                <input
                  v-model="addForm.name"
                  type="text"
                  placeholder="Full Name"
                  class="w-full px-3 py-2 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                />
              </div>

              <!-- Active + Welcome Mail -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div class="flex items-center">
                  <input type="checkbox" id="add-active" v-model="addForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                  <label for="add-active" class="ml-2 text-sm font-bold cursor-pointer">Active Account</label>
                </div>
                <div class="flex items-center">
                  <input type="checkbox" id="add-welcome" v-model="addForm.sendWelcome" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                  <label for="add-welcome" class="ml-2 text-sm font-bold cursor-pointer">Send Welcome Email</label>
                </div>
              </div>
            </div>

            <!-- Password Section (border-2 exact) -->
            <div class="border-2 border-brand-text p-4 space-y-3">
              <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
                <Icon name="key" :size="16" class="mr-2" />
                PASSWORD
              </h4>

              <div class="flex gap-2 items-center">
                <!-- Password -->
                <div class="relative flex-1">
                  <input
                    :type="showPw1 ? 'text' : 'password'"
                    v-model="addForm.password"
                    required
                    minlength="8"
                    placeholder="New password"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm pr-10"
                    @input="onAddPasswordInput"
                  />
                  <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary transition-colors" @click="showPw1 = !showPw1">
                    <Icon :name="showPw1 ? 'eye-off' : 'eye'" :size="16" />
                  </button>
                </div>
                <!-- Confirm -->
                <div class="relative flex-1">
                  <input
                    :type="showPw2 ? 'text' : 'password'"
                    v-model="addForm.passwordConfirm"
                    required
                    minlength="8"
                    placeholder="Repeat password"
                    class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm pr-10"
                    @input="onAddPasswordInput"
                  />
                  <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary transition-colors" @click="showPw2 = !showPw2">
                    <Icon :name="showPw2 ? 'eye-off' : 'eye'" :size="16" />
                  </button>
                </div>
                <!-- Generate (exact button style) -->
                <button type="button" @click="generateAddPassword"
                  class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white text-xs font-black px-3 h-10 border-2 border-brand-text cursor-pointer uppercase tracking-widest flex items-center gap-1 whitespace-nowrap flex-shrink-0 transition-colors">
                  <Icon name="wand-2" :size="14" />
                  GENERATE
                </button>
              </div>

              <p v-if="addPwdMismatch" class="text-xs text-red-600 font-bold">Passwords do not match</p>

              <!-- Strength Meter (exact structure) -->
              <div v-if="addForm.password" id="add-strengthMeter">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-black uppercase tracking-widest text-brand-text">STRENGTH</span>
                  <span class="text-xs font-bold uppercase tracking-wider" :style="{ color: addPwdStrength.color }">{{ addPwdStrength.label }}</span>
                </div>
                <div class="h-2 bg-gray-200 border border-brand-text overflow-hidden">
                  <div class="h-full transition-all duration-300" :style="{ width: addPwdStrength.pct + '%', background: addPwdStrength.color }"></div>
                </div>
              </div>
            </div>

            <!-- Advanced Settings (exact <details> from reference) -->
            <details class="border-2 border-brand-text">
              <summary class="px-4 py-3 cursor-pointer font-bold uppercase tracking-tight text-xs flex items-center hover:bg-gray-50 transition-colors">
                <Icon name="settings" :size="14" class="mr-2" />
                ADVANCED SETTINGS
                <Icon name="chevron-down" :size="14" class="ml-auto" />
              </summary>
              <div class="px-4 pb-4 pt-2 space-y-3 border-t-2 border-brand-text">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">QUOTA (MB)</label>
                    <input
                      v-model.number="addForm.quotaMB"
                      type="number"
                      min="0"
                      class="w-full px-3 py-2 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                    />
                    <p class="text-xs text-gray-500 mt-1">Storage limit in MB (0 = unlimited)</p>
                  </div>
                  <div class="flex items-center pt-5">
                    <input type="checkbox" id="add-smtp" v-model="addForm.smtpActive" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                    <label for="add-smtp" class="ml-2 text-sm font-bold cursor-pointer">SMTP Active (can send email)</label>
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALTERNATIVE EMAIL</label>
                  <input
                    v-model="addForm.emailOther"
                    type="email"
                    placeholder="other@example.com"
                    class="w-full px-3 py-2 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium transition-colors text-sm"
                  />
                </div>
              </div>
            </details>
          </form>
        </div>

        <!-- Footer Buttons (exact shadows and hover from reference) -->
        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
          <button type="button" @click="closeAdd"
            class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-6 py-3 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" :disabled="savingAdd" @click="submitAdd"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm disabled:opacity-60">
            <Icon name="save" :size="16" class="mr-2" />
            {{ savingAdd ? 'SAVING...' : 'SAVE EMAIL ACCOUNT' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ EDIT MODAL ══════════ -->
    <div v-if="showEdit" class="modal-overlay" @click.self="showEdit = false">
      <div class="modal-card" style="border: 4px solid #1e293b; max-width: 720px;">
        <!-- Blue header bar like the old template -->
        <div class="bg-brand-primary px-6 py-4 flex items-center justify-between flex-shrink-0" style="border-bottom: 2px solid #1e293b;">
          <h3 class="text-lg font-mono font-black uppercase tracking-tight text-white flex items-center">
            <Icon name="edit" :size="20" class="mr-2" />
            EDIT EMAIL ACCOUNT
          </h3>
          <button @click="showEdit = false" class="text-white hover:text-gray-200 transition-colors">
            <Icon name="x" :size="20" />
          </button>
        </div>
        <div class="modal-body">
          <div v-if="editError" class="mx-6 mt-4 bg-red-50 border-2 border-red-600 p-3 flex items-start">
            <Icon name="alert-circle" :size="16" class="text-red-600 mr-2 mt-0.5 flex-shrink-0" />
            <p class="text-sm text-red-700 font-medium">{{ editError }}</p>
          </div>

          <!-- EMAIL ACCOUNT Section (border-2 like old template) -->
          <div class="border-2 border-brand-text p-4 space-y-4">
            <h4 class="text-sm font-mono font-black uppercase tracking-tight flex items-center">
              <Icon name="mail" :size="16" class="mr-2" />
              EMAIL ACCOUNT
            </h4>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">EMAIL ADDRESS</label>
              <div class="h-10 px-3 flex items-center border-2 border-gray-300 bg-gray-50 font-mono text-sm font-medium text-gray-700">
                {{ editForm.username }}
              </div>
            </div>

            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">DISPLAY NAME</label>
              <input v-model="editForm.name" type="text" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm" placeholder="Full name" />
            </div>

            <div class="flex items-center">
              <input type="checkbox" id="edit-active" v-model="editForm.active" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
              <label for="edit-active" class="ml-2 text-sm font-bold cursor-pointer">Active Account</label>
            </div>
          </div>

          <!-- CHANGE PASSWORD (collapsible with border-2 like old template) -->
          <details class="border-2 border-brand-text">
            <summary class="px-4 py-3 cursor-pointer font-bold uppercase tracking-tight text-xs flex items-center hover:bg-gray-50 transition-colors">
              <Icon name="key" :size="14" class="mr-2" />
              CHANGE PASSWORD
              <Icon name="chevron-down" :size="14" class="ml-auto" />
            </summary>
            <div class="px-4 pb-4 pt-3 space-y-3 border-t-2 border-brand-text">
              <div class="flex gap-2 items-center">
                <div class="relative flex-1">
                  <input v-model="editForm.password" :type="showPw3 ? 'text' : 'password'" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm pr-10" placeholder="New password (min 8 chars)" @input="onEditPasswordInput" />
                  <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary" @click="showPw3 = !showPw3">
                    <Icon :name="showPw3 ? 'eye-off' : 'eye'" :size="16" />
                  </button>
                </div>
                <div class="relative flex-1">
                  <input v-model="editForm.passwordConfirm" :type="showPw4 ? 'text' : 'password'" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm pr-10" placeholder="Confirm new password" @input="onEditPasswordInput" />
                  <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-brand-primary" @click="showPw4 = !showPw4">
                    <Icon :name="showPw4 ? 'eye-off' : 'eye'" :size="16" />
                  </button>
                </div>
                <button type="button" class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white text-xs font-black px-3 h-10 border-2 border-brand-text cursor-pointer uppercase tracking-widest flex items-center gap-1 whitespace-nowrap flex-shrink-0 transition-colors" @click="generateEditPassword()">
                  <Icon name="wand-2" :size="14" class="mr-1" />
                  GENERATE
                </button>
              </div>

              <div v-if="editForm.password" class="pwd-strength">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-black uppercase tracking-widest text-brand-text">STRENGTH</span>
                  <span class="text-xs font-bold uppercase tracking-wider" :style="{ color: editPwdStrength.color }">{{ editPwdStrength.label }}</span>
                </div>
                <div class="h-2 bg-gray-200 border border-brand-text overflow-hidden">
                  <div class="h-full transition-all duration-300" :style="{ width: editPwdStrength.pct + '%', background: editPwdStrength.color }"></div>
                </div>
              </div>
              <div v-if="editPwdMismatch" class="text-xs text-red-600 font-bold">Passwords do not match</div>
            </div>
          </details>

          <!-- ADVANCED SETTINGS (collapsible with border-2 like old template, open by default) -->
          <details class="border-2 border-brand-text" open>
            <summary class="px-4 py-3 cursor-pointer font-bold uppercase tracking-tight text-xs flex items-center hover:bg-gray-50 transition-colors">
              <Icon name="settings" :size="14" class="mr-2" />
              ADVANCED SETTINGS
              <Icon name="chevron-down" :size="14" class="ml-auto" />
            </summary>
            <div class="px-4 pb-4 pt-2 space-y-3 border-t-2 border-brand-text">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">QUOTA (MB)</label>
                  <input v-model.number="editForm.quotaMB" type="number" min="0" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm" />
                  <p class="text-[10px] text-gray-400 mt-1">Storage limit in MB (0 = unlimited)</p>
                </div>
                <div class="flex items-center pt-5">
                  <input type="checkbox" id="edit-smtp" v-model="editForm.smtpActive" class="w-5 h-5 border-2 border-brand-text cursor-pointer" />
                  <label for="edit-smtp" class="ml-2 text-sm font-bold cursor-pointer">SMTP Active (can send email)</label>
                </div>
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-brand-text mb-1">ALTERNATIVE EMAIL</label>
                <input v-model="editForm.emailOther" type="email" class="w-full h-10 px-3 border-2 border-brand-text focus:border-brand-primary focus:outline-none font-medium text-sm" placeholder="other@example.com" />
              </div>
            </div>
          </details>
        </div>
        <!-- Footer buttons matching old template style -->
        <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t-2 border-brand-text flex-shrink-0 bg-white">
          <button type="button" @click="showEdit = false"
            class="bg-white hover:bg-gray-50 text-brand-text border-2 border-brand-text font-black px-6 py-3 shadow-[2px_2px_0px_#1E293B] transition-all hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[3px_3px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
            <Icon name="x" :size="16" class="mr-2" />
            CANCEL
          </button>
          <button type="button" @click="submitEdit" :disabled="savingEdit"
            class="bg-brand-primary hover:bg-white hover:text-brand-primary text-white border-2 border-brand-text font-black px-6 py-3 shadow-[3px_3px_0px_#1E293B] transition-all hover:-translate-x-1 hover:-translate-y-1 hover:shadow-[4px_4px_0px_#1E293B] active:translate-x-0 active:translate-y-0 active:shadow-none cursor-pointer uppercase tracking-widest flex items-center text-sm">
            <Icon name="save" :size="16" class="mr-2" />
            {{ savingEdit ? 'SAVING...' : 'UPDATE EMAIL ACCOUNT' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══════════ DELETE CONFIRM ══════════ -->
    <BrutalModal v-model="showDeleteConfirm" title="CONFIRM DELETE" size="sm" danger>
      <p class="confirm-text">
        Are you sure you want to delete mailbox<br />
        <strong>{{ deleteTarget?.username }}</strong>?<br />
        <span class="confirm-sub">This action cannot be undone.</span>
      </p>

      <template #footer>
        <button class="btn-cancel" @click="showDeleteConfirm = false">CANCEL</button>
        <button class="btn-danger" :disabled="deletingRow" @click="submitDelete">
          <Icon name="trash-2" :size="14" style="margin-right:6px;vertical-align:middle" />
          {{ deletingRow ? 'DELETING...' : 'DELETE' }}
        </button>
      </template>
    </BrutalModal>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import axios from 'axios'
import { useToastStore } from '../stores/toast'
import { calcStrength, generatePassword } from '../utils/password'

const toast = useToastStore()

const QUOTA_MULTIPLIER = 1048576

interface Mailbox {
  username: string
  domain: string
  name: string
  quota: number
  active: boolean
  smtp_active: boolean
  email_other: string
  modified: string
  created: string
}

interface Domain { domain: string }

const allMailboxes = ref<Mailbox[]>([])
const domains = ref<Domain[]>([])
const loading = ref(true)
const error = ref('')

const search = ref('')
const rowsPerPage = ref(15)
const currentPage = ref(1)
const sortKey = ref('username')
const sortDir = ref<'asc' | 'desc'>('asc')
const domainFilter = ref('')

const columns = [
  { key: 'username', label: 'EMAIL' },
  { key: 'name',     label: 'NAME' },
  { key: 'domain',   label: 'DOMAIN' },
  { key: 'quota',    label: 'QUOTA' },
  { key: 'active',   label: 'ACTIVE' },
  { key: 'modified', label: 'MODIFIED' },
]

function sortBy(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

const filteredRows = computed(() => {
  let rows = allMailboxes.value
  if (domainFilter.value) rows = rows.filter(r => r.domain === domainFilter.value)
  const q = search.value.toLowerCase()
  if (q) rows = rows.filter(r =>
    r.username.toLowerCase().includes(q) ||
    r.name.toLowerCase().includes(q) ||
    r.domain.toLowerCase().includes(q)
  )
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
watch([search, rowsPerPage, domainFilter], () => { currentPage.value = 1 })

function formatDate(ts: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('pt-BR', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function load() {
  loading.value = true; error.value = ''
  try {
    const [mbRes, domRes] = await Promise.all([
      axios.get(`${API_BASE}/mailboxes`),
      axios.get(`${API_BASE}/domains`),
    ])
    allMailboxes.value = mbRes.data?.data ?? []
    domains.value = domRes.data?.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load mailboxes'
  } finally { loading.value = false }
}
onMounted(load)

// ─── Add modal ───
const showAdd = ref(false)
const savingAdd = ref(false)
const addError = ref('')
const showPw1 = ref(false)
const showPw2 = ref(false)
const addPwdMismatch = ref(false)

const addForm = ref({
  localPart: '', domain: '', name: '', quotaMB: 1024,
  active: true, smtpActive: true, sendWelcome: false,
  emailOther: '', password: '', passwordConfirm: '',
})

const addPwdStrength = computed(() => calcStrength(addForm.value.password))

function onAddPasswordInput() {
  addPwdMismatch.value = !!(addForm.value.passwordConfirm && addForm.value.password !== addForm.value.passwordConfirm)
}

function generateAddPassword() {
  generatePassword(addForm.value)
  showPw1.value = true
  showPw2.value = true
  onAddPasswordInput()
}

function openAdd() {
  addForm.value = {
    localPart: '', domain: domains.value[0]?.domain || '', name: '', quotaMB: 1024,
    active: true, smtpActive: true, sendWelcome: false,
    emailOther: '', password: '', passwordConfirm: '',
  }
  addError.value = ''; showPw1.value = false; showPw2.value = false; addPwdMismatch.value = false
  showAdd.value = true
}

function closeAdd() {
  showAdd.value = false
  // reset like the reference
  addForm.value = {
    localPart: '', domain: '', name: '', quotaMB: 1024,
    active: true, smtpActive: true, sendWelcome: false,
    emailOther: '', password: '', passwordConfirm: '',
  }
  addError.value = ''
  showPw1.value = false
  showPw2.value = false
  addPwdMismatch.value = false
}

function onLocalPartInput(e: Event) {
  const val = (e.target as HTMLInputElement).value.toLowerCase()
  addForm.value.localPart = val
}

function updateEmailPreview() {
  // no-op, template uses direct binding on addForm
}

async function submitAdd() {
  addError.value = ''
  const f = addForm.value
  if (!f.localPart || f.localPart.length < 4) { addError.value = 'Username must be at least 4 characters'; return }
  if (!f.domain) { addError.value = 'Please select a domain'; return }
  if (!f.password || f.password.length < 8) { addError.value = 'Password must be at least 8 characters'; return }
  if (f.password !== f.passwordConfirm) { addError.value = 'Passwords do not match'; return }
  savingAdd.value = true
  try {
    await axios.post(`${API_BASE}/mailboxes`, {
      local_part: f.localPart.toLowerCase().trim(),
      domain: f.domain,
      name: f.name,
      password: f.password,
      quota: f.quotaMB > 0 ? f.quotaMB * QUOTA_MULTIPLIER : 0,
      active: f.active,
      smtp_active: f.smtpActive,
      send_welcome: f.sendWelcome,
      email_other: f.emailOther,
    })
    closeAdd()
    toast.success(`Mailbox ${f.localPart}@${f.domain} created successfully`)
    await load()
  } catch (e: any) {
    addError.value = e?.response?.data?.error?.message || 'Failed to create mailbox'
    toast.error(addError.value)
  } finally { savingAdd.value = false }
}

// ─── Edit modal ───
const showEdit = ref(false)
const savingEdit = ref(false)
const editError = ref('')
const showPw3 = ref(false)
const showPw4 = ref(false)
const editPwdMismatch = ref(false)

const editForm = ref({
  username: '', name: '', quotaMB: 0,
  active: true, smtpActive: true, emailOther: '',
  password: '', passwordConfirm: '',
})

const editPwdStrength = computed(() => calcStrength(editForm.value.password))

function onEditPasswordInput() {
  editPwdMismatch.value = !!(editForm.value.passwordConfirm && editForm.value.password !== editForm.value.passwordConfirm)
}

function generateEditPassword() {
  generatePassword(editForm.value)
  showPw3.value = true
  showPw4.value = true
  onEditPasswordInput()
}

async function openEdit(row: Mailbox) {
  editError.value = ''; showPw3.value = false; showPw4.value = false; editPwdMismatch.value = false
  try {
    const res = await axios.get(`${API_BASE}/mailboxes/${encodeURIComponent(row.username)}`)
    const mb = res.data?.data
    editForm.value = {
      username: mb.username,
      name: mb.name || '',
      quotaMB: mb.quota,
      active: mb.active,
      smtpActive: mb.smtp_active,
      emailOther: mb.email_other || '',
      password: '',
      passwordConfirm: '',
    }
    showEdit.value = true
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load mailbox'
  }
}

async function submitEdit() {
  editError.value = ''
  const f = editForm.value
  const payload: any = {
    name: f.name,
    quota: f.quotaMB,
    active: f.active,
    smtp_active: f.smtpActive,
    email_other: f.emailOther,
  }
  if (f.password) {
    if (f.password.length < 8) { editError.value = 'Password must be at least 8 characters'; return }
    if (f.password !== f.passwordConfirm) { editError.value = 'Passwords do not match'; return }
    payload.change_password = true
    payload.password = f.password
    payload.password_confirm = f.passwordConfirm
  }
  savingEdit.value = true
  try {
    await axios.put(`${API_BASE}/mailboxes/${encodeURIComponent(f.username)}`, payload)
    showEdit.value = false
    toast.success(`Mailbox ${f.username} updated successfully`)
    await load()
  } catch (e: any) {
    editError.value = e?.response?.data?.error?.message || 'Failed to update mailbox'
    toast.error(editError.value)
  } finally { savingEdit.value = false }
}

// ─── Delete ───
const showDeleteConfirm = ref(false)
const deletingRow = ref(false)
const deleteTarget = ref<Mailbox | null>(null)

function confirmDelete(row: Mailbox) { deleteTarget.value = row; showDeleteConfirm.value = true }

async function submitDelete() {
  if (!deleteTarget.value) return
  deletingRow.value = true
  try {
    const username = deleteTarget.value.username
    await axios.delete(`${API_BASE}/mailboxes/${encodeURIComponent(username)}`)
    showDeleteConfirm.value = false; deleteTarget.value = null
    toast.success(`Mailbox ${username} deleted successfully`)
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Failed to delete mailbox'
    error.value = msg
    toast.error(msg)
    showDeleteConfirm.value = false
  } finally { deletingRow.value = false }
}

function onFilterChange() {
  // filter already applied in filteredRows
}
</script>

<style scoped>
.dom-page { background: #ebf2fe; padding: 24px 28px 40px; }

.dom-header {
  display: flex; justify-content: space-between; align-items: flex-start;
  margin-bottom: 20px;
}
.dom-title { font-size: 28px; font-weight: 900; color: #1e293b; letter-spacing: -0.5px; line-height: 1; font-family: monospace; text-transform: uppercase; }
.dom-subtitle { font-size: 10px; color: #94a3b8; letter-spacing: 0.8px; margin-top: 6px; text-transform: uppercase; font-weight: 700; }

/* .btn-primary, .error-banner, .table-card now centralized in global style.css (Tailwind-friendly) */

.table-topbar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 14px; border-bottom: 1px solid #e2e8f0; gap: 12px; flex-wrap: wrap;
}
.controls-left { display: flex; align-items: center; gap: 8px; }
.controls-right { display: flex; align-items: center; gap: 8px; }
.per-page-wrap { display: flex; align-items: center; gap: 6px; }
.ctrl-select {
  border: 1px solid #d1d5db; padding: 4px 6px; font-size: 13px; color: #374151;
  background: #fff; border-radius: 0; outline: none;
}
.ctrl-label { font-size: 12px; color: #64748b; font-weight: 500; }
.search-input {
  border: 1px solid #d1d5db; padding: 4px 8px; font-size: 13px; color: #374151;
  width: 200px; outline: none; border-radius: 0;
}
.search-input:focus { border-color: #3b82f6; }

.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
.table-head-row { background: #3b82f6; }
.table-th {
  color: #fff; font-weight: 700; letter-spacing: 0.4px; padding: 10px;
  text-align: left; cursor: pointer; white-space: nowrap; user-select: none;
  font-size: 12px;
}
.table-th:hover { background: #2563eb; }
.sort-arrows { margin-left: 4px; font-size: 9px; opacity: .5; }
.sort-active { opacity: 1 !important; }

.table-row:nth-child(even) { background: #f8fafc; }
.table-row:hover { background: #eff6ff; }
.table-td { padding: 6px 10px; color: #374151; border-bottom: 1px solid #f1f5f9; font-size: 12px; }
.td-link { color: #1e293b; font-weight: 600; }
.cell-with-icon { display: flex; align-items: center; gap: 6px; }
.row-icon { color: #3b82f6; flex-shrink: 0; }

.badge-yes { background: #dcfce7; color: #16a34a; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.badge-no  { background: #fee2e2; color: #dc2626; padding: 2px 8px; font-size: 11px; font-weight: 700; }

.actions-td { display: flex; gap: 6px; align-items: center; }
.act-btn {
  padding: 4px 10px; font-size: 10px; font-weight: 800; cursor: pointer;
  border: 1px solid #1e293b; letter-spacing: 0.4px; border-radius: 0;
  display: inline-flex; align-items: center; transition: all .12s;
  box-shadow: 1px 1px 0 #1e293b; text-transform: uppercase;
}
.act-btn:hover { transform: translate(-0.5px, -0.5px); }
.act-btn:active { transform: translate(0,0); box-shadow: none; }
.act-edit { background: #3b82f6; color: #fff; }
.act-edit:hover { background: #fff; color: #3b82f6; }
.act-del  { background: #ef4444; color: #fff; }
.act-del:hover { background: #fff; color: #ef4444; }

.table-loading, .table-empty { text-align: center; padding: 28px; color: #94a3b8; font-size: 13px; }

.table-footer {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 14px; border-top: 1px solid #e2e8f0;
}
.showing-text { font-size: 12.5px; color: #64748b; }
.pagination { display: flex; gap: 3px; }
.pg-btn { height: 28px; padding: 0 10px; font-size: 10px; font-weight: 700; color: #374151; background: #fff; border: 1px solid #d1d5db; border-radius: 0; cursor: pointer; letter-spacing: 0.4px; text-transform: uppercase; white-space: nowrap; }
.pg-btn:hover:not(:disabled) { border-color: #1e293b; color: #1e293b; background: #f8fafc; }
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-active { background: #3b82f6 !important; color: #fff !important; border-color: #3b82f6 !important; }
</style>
