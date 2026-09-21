<template>
  <div class="filter-bar">
    <template v-for="field in fields" :key="fieldKey(field)">
      <el-input
        v-if="field.type === 'input'"
        :model-value="modelValue[field.key]"
        :placeholder="field.placeholder"
        clearable
        @update:model-value="update(field.key, $event)"
        @keyup.enter="emitSearch"
        @clear="emitSearchIf(field)"
      />

      <el-select
        v-else-if="field.type === 'select'"
        :model-value="modelValue[field.key]"
        :placeholder="field.placeholder"
        clearable
        @update:model-value="update(field.key, $event)"
        @change="emitSearchIf(field)"
      >
        <el-option
          v-for="option in resolveOptions(field)"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>

      <el-date-picker
        v-else-if="field.type === 'daterange'"
        :model-value="dateRangeValue(field)"
        type="daterange"
        value-format="YYYY-MM-DD"
        range-separator="至"
        :start-placeholder="field.startPlaceholder"
        :end-placeholder="field.endPlaceholder"
        @update:model-value="updateDateRange(field, $event)"
        @change="emitSearch"
      />

      <el-checkbox
        v-else-if="field.type === 'checkbox'"
        :model-value="modelValue[field.key]"
        @update:model-value="update(field.key, $event)"
        @change="emitSearchIf(field)"
      >
        {{ field.label }}
      </el-checkbox>
    </template>

    <el-button type="primary" :icon="Search" @click="emitSearch">查询</el-button>
    <el-button :icon="RefreshLeft" @click="emitReset">重置</el-button>
  </div>
</template>

<script setup>
import { Search, RefreshLeft } from '@element-plus/icons-vue'
import { dictOptions } from '@/constants/dict'

const props = defineProps({
  // 筛选字段声明列表(见 constants/filters.js)。
  fields: { type: Array, default: () => [] },
  // 与声明 key 对应的查询对象, 由 useListPage 提供。
  modelValue: { type: Object, required: true },
  // 供 select 字段解析动态选项的字典 store (useDictStore)。
  dictStore: { type: Object, default: null },
})

const emit = defineEmits(['update:modelValue', 'search', 'reset'])

function update(key, value) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function fieldKey(field) {
  return field.type === 'daterange' ? `${field.startKey}-${field.endKey}` : field.key
}

// 日期区间在 query 中以两个字符串字段存在, el-date-picker 需要数组, 这里做双向转换。
function dateRangeValue(field) {
  const start = props.modelValue[field.startKey]
  const end = props.modelValue[field.endKey]
  return start || end ? [start, end] : []
}

function updateDateRange(field, range) {
  emit('update:modelValue', {
    ...props.modelValue,
    [field.startKey]: range?.[0] ?? '',
    [field.endKey]: range?.[1] ?? '',
  })
}

function emitSearch() {
  emit('search')
}

function emitReset() {
  emit('reset')
}

function emitSearchIf(field) {
  if (field.searchOnChange) emit('search')
}

// select 选项统一归一化为 [{value,label}], 支持:
//  - 静态字典对象(RUN_STATUS / FAULT_STATUS ...)
//  - 字符串数组(后端 meta 返回的 fault_types、repairmen ...)
//  - 函数(传入 dictStore 后返回上述两种之一)
function resolveOptions(field) {
  let options = field.options
  if (typeof options === 'function') {
    options = options(props.dictStore)
  }
  if (Array.isArray(options)) {
    return options.map((item) =>
      typeof item === 'string' ? { value: item, label: item } : item,
    )
  }
  return dictOptions(options ?? {})
}
</script>
