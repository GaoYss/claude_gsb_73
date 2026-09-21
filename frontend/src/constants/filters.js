// 列表页筛选字段的声明式描述。
//
// 每个列表页只需要"声明有哪些条件", 不再手写 el-input/el-select/日期区间,
// 也不再各自维护默认值、日期同步与重置逻辑。
// 字段的 key 与后端 form 参数名一一对应, 由 useListPage 统一构建查询对象,
// 从根上消除"前端可选项"与"后端允许取值"靠人工对齐的问题。

// 关键词输入框, 固定映射到后端的 keyword 参数。
export function keywordField(placeholder) {
  return { type: 'input', key: 'keyword', placeholder, defaultValue: '', clearable: true }
}

// 普通文本输入框。
export function inputField(key, placeholder) {
  return { type: 'input', key, placeholder, defaultValue: '', clearable: true }
}

// 下拉选择。
// options 由 resolveFieldOptions 解析, 支持静态字典 / 静态数组 / 字典 store 三种来源。
// searchOnChange 为 true 时选择后立即查询(故障、维修页的既有行为)。
export function selectField(key, placeholder, options, searchOnChange = false) {
  return { type: 'select', key, placeholder, options, defaultValue: '', clearable: true, searchOnChange }
}

// 日期区间, 映射为两个后端参数 startKey/endKey(半开区间由后端统一处理)。
export function dateRangeField({
  startKey = 'start_date',
  endKey = 'end_date',
  startPlaceholder = '开始日期',
  endPlaceholder = '结束日期',
} = {}) {
  return { type: 'daterange', startKey, endKey, startPlaceholder, endPlaceholder }
}

// 布尔复选框, 默认值 false。
export function checkboxField(key, label, searchOnChange = false) {
  return { type: 'checkbox', key, label, defaultValue: false, searchOnChange }
}

// queryDefaults 依据字段声明生成查询对象的初始(重置)值。
// 分页字段不在此处理(由 useListPage 统一提供)。
export function queryDefaults(fields) {
  const defaults = {}
  for (const field of fields) {
    if (field.type === 'daterange') {
      defaults[field.startKey] = ''
      defaults[field.endKey] = ''
      continue
    }
    defaults[field.key] = field.defaultValue ?? (field.type === 'checkbox' ? false : '')
  }
  return defaults
}

