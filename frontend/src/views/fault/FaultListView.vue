<template>
  <div class="page">
    <PageHeader title="故障登记" description="受理路灯故障上报, 跟踪从登记到闭环的完整状态流转">
      <el-button :icon="Refresh" @click="load">刷新</el-button>
      <el-button type="primary" :icon="Plus" @click="openCreate">登记故障</el-button>
    </PageHeader>

    <el-card shadow="never">
      <FilterBar
        :fields="faultFilterFields"
        :model-value="query"
        :dict-store="dictStore"
        @update:model-value="patchQuery"
        @search="search"
        @reset="reset"
      />
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="fault_no" label="故障单号" width="140" fixed="left" />
        <el-table-column prop="lamp_code" label="路灯编号" width="110" />
        <el-table-column prop="road_name" label="所在道路" min-width="120" show-overflow-tooltip />
        <el-table-column prop="fault_type" label="故障类型" width="110" />
        <el-table-column label="紧急程度" width="100">
          <template #default="{ row }"><StatusTag :dict="FAULT_LEVEL" :value="row.fault_level" /></template>
        </el-table-column>
        <el-table-column label="处理状态" width="100">
          <template #default="{ row }"><StatusTag :dict="FAULT_STATUS" :value="row.status" /></template>
        </el-table-column>
        <el-table-column label="来源" width="100">
          <template #default="{ row }">{{ dictLabel(FAULT_SOURCE, row.source) }}</template>
        </el-table-column>
        <el-table-column prop="reporter" label="上报人" width="100" />
        <el-table-column label="上报时间" width="150">
          <template #default="{ row }">{{ formatDateTime(row.reported_at) }}</template>
        </el-table-column>
        <el-table-column label="维修次数" width="90" align="center">
          <template #default="{ row }">{{ row.repair_count }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="isOpen(row)" link type="warning" @click="openRepair(row)">维修录入</el-button>
            <el-button v-if="row.status !== 'closed'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status !== 'closed'" link type="info" @click="handleClose(row)">关闭</el-button>
            <el-button v-if="row.status === 'closed'" link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <FaultFormDialog
      v-model="formVisible"
      :model="editing"
      :preset-lamp="presetLamp"
      :fault-type-options="faultTypeOptions"
      @saved="handleSaved"
    />
    <FaultDetailDrawer v-model="detailVisible" :fault-id="activeFaultId" />
    <RepairFormDialog v-model="repairVisible" :fault="repairTarget" @saved="handleSaved" />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusTag from '@/components/common/StatusTag.vue'
import DataPagination from '@/components/common/DataPagination.vue'
import FilterBar from '@/components/common/FilterBar.vue'
import FaultFormDialog from './components/FaultFormDialog.vue'
import FaultDetailDrawer from './components/FaultDetailDrawer.vue'
import RepairFormDialog from '@/views/repair/components/RepairFormDialog.vue'
import { faultApi } from '@/api/fault'
import { lampApi } from '@/api/lamp'
import { useDictStore } from '@/stores/dict'
import { FAULT_LEVEL, FAULT_SOURCE, FAULT_STATUS, dictLabel, isFaultOpen } from '@/constants/dict'
import { faultFilterFields } from '@/constants/listSchemas'
import { formatDateTime } from '@/utils/format'
import { useListPage } from '@/composables/useListPage'

const route = useRoute()
const router = useRouter()
const dictStore = useDictStore()

const { loading, rows, total, query, load, search, reset, changePage, changePageSize, patchQuery } =
  useListPage(faultApi.list, faultFilterFields)

const faultTypeOptions = computed(() => dictStore.faultMeta.fault_types)

const formVisible = ref(false)
const detailVisible = ref(false)
const repairVisible = ref(false)
const editing = ref(null)
const presetLamp = ref(null)
const repairTarget = ref(null)
const activeFaultId = ref(null)

const isOpen = (row) => isFaultOpen(row.status)

function openCreate() {
  editing.value = null
  formVisible.value = true
}

function openEdit(row) {
  editing.value = { ...row }
  formVisible.value = true
}

function openDetail(row) {
  activeFaultId.value = row.id
  detailVisible.value = true
}

function openRepair(row) {
  repairTarget.value = { ...row }
  repairVisible.value = true
}

async function handleClose(row) {
  try {
    const { value } = await ElMessageBox.prompt('请输入关闭说明 (例如: 现场复核通过 / 误报作废)', `关闭故障 ${row.fault_no}`, {
      confirmButtonText: '确认关闭',
      cancelButtonText: '取消',
      inputPlaceholder: '关闭说明',
    })
    await faultApi.close(row.id, { remark: value ?? '' })
    ElMessage.success('故障已关闭')
    load()
  } catch (error) {
    // 用户取消或请求失败, 提示由拦截器处理
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除故障 ${row.fault_no} ? 已产生维修记录的故障无法删除。`, '删除确认', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
    })
  } catch (error) {
    return
  }

  try {
    await faultApi.remove(row.id)
    ElMessage.success('故障已删除')
    load()
  } catch (error) {
    // 错误提示由请求拦截器统一处理
  }
}

function handleSaved() {
  load()
  dictStore.refreshAll().catch(() => {})
}

// 支持从路灯台账页携带 lamp_id 直接发起故障登记。
async function applyRouteQuery() {
  const lampId = Number(route.query.lamp_id)
  if (!lampId) return

  try {
    presetLamp.value = await lampApi.detail(lampId, { silent: true })
    editing.value = null
    formVisible.value = true
  } catch (error) {
    presetLamp.value = null
  }
  router.replace({ path: '/faults' })
}

onMounted(async () => {
  dictStore.ensureLoaded().catch(() => {})
  await applyRouteQuery()
})
</script>
