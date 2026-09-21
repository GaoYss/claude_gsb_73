// 各列表页的筛选条件声明。
//
// 页面只声明"需要哪些条件", 渲染(FilterBar)、默认值/重置(queryDefaults)、
// 分页与请求拼装(useListPage)全部由通用实现承担。
// 字段 key 与后端 form 参数名严格一致; 新增一个筛选维度时,
// 只需在对应 schema 增加一项并确保后端接受同名参数, 不必再改多个页面。
import {
  FAULT_LEVEL,
  FAULT_STATUS,
  REPAIR_RESULT,
  REPAIR_STATUS,
  RUN_STATUS,
} from '@/constants/dict'
import { checkboxField, dateRangeField, keywordField, selectField } from '@/constants/filters'

// 路灯台账页(/lamps)。
export const lampFilterFields = [
  keywordField('编号 / 名称 / 道路 / 地址'),
  selectField('road_name', '所在道路', (store) => store?.lampOptions?.roads ?? []),
  selectField('lamp_type', '灯具类型', (store) => store?.lampOptions?.lamp_types ?? []),
  selectField('run_status', '运行状态', RUN_STATUS),
]

// 故障登记页(/faults)。
export const faultFilterFields = [
  keywordField('故障单号 / 路灯编号 / 道路 / 描述'),
  selectField('status', '处理状态', FAULT_STATUS, true),
  selectField('fault_type', '故障类型', (store) => store?.faultMeta?.fault_types ?? [], true),
  selectField('fault_level', '紧急程度', FAULT_LEVEL, true),
  dateRangeField({
    startPlaceholder: '上报开始日期',
    endPlaceholder: '上报结束日期',
  }),
  checkboxField('only_open', '仅看未闭环', true),
]

// 维修记录页(/repairs)。
export const repairFilterFields = [
  keywordField('维修单号 / 故障单号 / 路灯编号 / 维修人员'),
  selectField('status', '维修状态', REPAIR_STATUS, true),
  selectField('result', '维修结果', REPAIR_RESULT, true),
  selectField('repairman', '维修人员', (store) => store?.repairMeta?.repairmen ?? [], true),
  dateRangeField({
    startPlaceholder: '开工开始日期',
    endPlaceholder: '开工结束日期',
  }),
]

// 维修状态查询页(/status/lamps)。
// 台账维度的条件与路灯台账页保持一致(后端共用 lamp.LedgerConditions),
// 额外的 only_open 是该读模型特有的存在性条件。
export const statusLampFilterFields = [
  keywordField('路灯编号 / 名称 / 道路 / 地址'),
  selectField('road_name', '所在道路', (store) => store?.lampOptions?.roads ?? []),
  selectField('run_status', '运行状态', RUN_STATUS),
  checkboxField('only_open', '仅看有未闭环故障'),
]