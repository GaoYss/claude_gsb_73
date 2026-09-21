<template>
  <div class="page">
    <PageHeader title="维修状态查询" description="按路灯维度查看当前故障与最近一次维修进展, 快速定位滞留工单">
      <el-button :icon="Refresh" @click="load">刷新</el-button>
    </PageHeader>

    <el-card shadow="never">
      <FilterBar
        :fields="statusLampFilterFields"
        :model-value="query"
        :dict-store="dictStore"
        @update:model-value="patchQuery"
        @search="search"
        @reset="reset"
      />
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="lamp_code" label="路灯编号" width="120" fixed="left" />
        <el-table-column prop="lamp_name" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="road_name" label="所在道路" min-width="120" show-overflow-tooltip />
        <el-table-column label="运行状态" width="100">
          <template #default="{ row }"><StatusTag :dict="RUN_STATUS" :value="row.run_status" /></template>
        </el-table-column>
        <el-table-column label="故障统计" width="150">
          <template #default="{ row }">
            <span>未闭环 {{ row.open_fault_count }} / 累计 {{ row.total_fault_count }}</span>
          </template>
        </el-table-column>
        <el-table-column label="当前故障" min-width="230">
          <template #default="{ row }">
            <template v-if="row.current_fault_no">
              <div class="cell-main">
                <el-link type="primary" @click="goTrack({ fault_no: row.current_fault_no })">
                  {{ row.current_fault_no }}
                </el-link>
                <StatusTag :dict="FAULT_STATUS" :value="row.current_fault_status" />
              </div>
              <div class="cell-sub text-muted">
                {{ row.current_fault_type }} · 上报于 {{ formatDateTime(row.current_fault_reported_at) }}
              </div>
            </template>
            <span v-else class="text-muted">无未闭环故障</span>
          </template>
        </el-table-column>
        <el-table-column label="最近维修" min-width="240">
          <template #default="{ row }">
            <template v-if="row.latest_repair_no">
              <div class="cell-main">
                <span>{{ row.latest_repair_no }}</span>
                <StatusTag :dict="REPAIR_STATUS" :value="row.latest_repair_status" />
              </div>
              <div class="cell-sub text-muted">
                {{ row.latest_repairman || '未指派' }} · {{ dictLabel(REPAIR_RESULT, row.latest_repair_result) }} ·
                {{ formatDateTime(row.latest_repaired_at) }}
              </div>
            </template>
            <span v-else class="text-muted">暂无维修记录</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="goTrack({ lamp_code: row.lamp_code })">进度追踪</el-button>
          </template>
        </el-table-column>
      </el-table>
      <DataPagination
        :page="query.page"
        :page-size="query.page_size"
        :total="total"
        @page-change="changePage"
        @size-change="changePageSize"
      />
    </el-card>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusTag from '@/components/common/StatusTag.vue'
import DataPagination from '@/components/common/DataPagination.vue'
import FilterBar from '@/components/common/FilterBar.vue'
import { statusApi } from '@/api/status'
import { useDictStore } from '@/stores/dict'
import { FAULT_STATUS, REPAIR_RESULT, REPAIR_STATUS, RUN_STATUS, dictLabel } from '@/constants/dict'
import { statusLampFilterFields } from '@/constants/listSchemas'
import { formatDateTime } from '@/utils/format'
import { useListPage } from '@/composables/useListPage'

const router = useRouter()
const dictStore = useDictStore()

const { loading, rows, total, query, load, search, reset, changePage, changePageSize, patchQuery } =
  useListPage(statusApi.lamps, statusLampFilterFields)

function goTrack(params) {
  router.push({ path: '/status/track', query: params })
}

onMounted(() => {
  dictStore.ensureLoaded().catch(() => {})
})
</script>

<style scoped>
.cell-main {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cell-sub {
  font-size: 12px;
  margin-top: 2px;
}
</style>
