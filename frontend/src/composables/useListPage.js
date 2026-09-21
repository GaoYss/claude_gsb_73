import { onMounted, reactive, ref } from 'vue'
import { queryDefaults } from '@/constants/filters'

// 列表页通用逻辑: 分页查询、条件重置、翻页。
//
// 页面通过 fields 声明需要的筛选条件(见 constants/filters.js),
// 本组合式函数据此构建查询对象与重置基线:
//   - 新增/删除一个筛选维度只需改字段声明, 不必再动分页与重置逻辑;
//   - 日期区间直接以 start_*/end_* 两个字段存在于 query 中,
//     无需各页面再维护一个 dateRange ref 和同步函数。
export function useListPage(fetcher, fields = [], options = {}) {
  const { immediate = true, overrides = {} } = options
  const loading = ref(false)
  const rows = ref([])
  const total = ref(0)

  const buildDefault = () => ({ page: 1, page_size: 10, ...queryDefaults(fields), ...overrides })

  const query = reactive(buildDefault())

  async function load() {
    loading.value = true
    try {
      const data = await fetcher({ ...query })
      rows.value = data?.items ?? []
      total.value = data?.total ?? 0
    } catch (error) {
      rows.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  function search() {
    query.page = 1
    return load()
  }

  function reset() {
    Object.assign(query, buildDefault())
    return load()
  }

  function changePage(page) {
    query.page = page
    return load()
  }

  function changePageSize(pageSize) {
    query.page = 1
    query.page_size = pageSize
    return load()
  }

  // 供声明式 FilterBar 回写整个查询对象(仅做字段合并, 保持 query 为同一 reactive 引用)。
  function patchQuery(next) {
    Object.assign(query, next)
  }

  if (immediate) {
    onMounted(load)
  }

  return { loading, rows, total, query, load, search, reset, changePage, changePageSize, patchQuery }
}
