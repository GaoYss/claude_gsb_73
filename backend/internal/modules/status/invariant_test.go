package status_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"streetlight/internal/modules/lamp"
	"streetlight/internal/modules/status"
)

// TestLampListAndStatusListShareConditions 锁定核心不变量:
// 路灯台账列表(/lamps)与维修状态读模型(/status/lamps)在相同台账条件下
// (keyword / road_name / lamp_type / run_status), 必须得到相同的总数与行顺序。
// 两者唯一允许的差异是行的附加字段(故障统计/维修进展), 归属集合必须一致。
func TestLampListAndStatusListShareConditions(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	h.createLamp(t, "LD-I-001", "中山路")
	h.createLamp(t, "LD-I-002", "中山路")
	h.createLamp(t, "LD-I-003", "解放路")
	h.createLamp(t, "LD-I-004", "解放路")

	// 让一盏路灯产生未闭环故障, 运行状态变为 fault, 但不影响台账条件口径。
	faulty := h.createLamp(t, "LD-I-005", "中山路")
	h.createFault(t, faulty.ID, "灯不亮")

	cases := []struct {
		name    string
		keyword string
		road    string
		status_ string
	}{
		{name: "无条件"},
		{name: "按道路-中山路", road: "中山路"},
		{name: "按道路-解放路", road: "解放路"},
		{name: "按编号关键词", keyword: "LD-I-00"},
		{name: "按精确编号", keyword: "LD-I-003"},
		{name: "按运行状态-正常", status_: lamp.RunStatusNormal},
		{name: "按运行状态-故障", status_: lamp.RunStatusFault},
		{name: "道路+状态", road: "中山路", status_: lamp.RunStatusNormal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ledgerItems, ledgerTotal, _, err := h.lamps.List(ctx, lamp.ListQuery{
				Keyword:   tc.keyword,
				RoadName:  tc.road,
				RunStatus: tc.status_,
			})
			require.NoError(t, err)

			statusRows, statusTotal, _, err := h.status.Lamps(ctx, status.LampQuery{
				Keyword:   tc.keyword,
				RoadName:  tc.road,
				RunStatus: tc.status_,
			})
			require.NoError(t, err)

			// 不变量 1: 同条件总数一致。
			require.Equal(t, ledgerTotal, statusTotal, "两个读模型总数必须一致")
			require.Len(t, statusRows, len(ledgerItems), "当前页行数必须一致")

			// 不变量 2: 同条件同分页下, 行顺序(以路灯编号代表)逐行一致。
			for i := range ledgerItems {
				require.Equal(t, ledgerItems[i].Code, statusRows[i].LampCode,
					"第 %d 行归属不一致: 台账=%s 状态=%s", i, ledgerItems[i].Code, statusRows[i].LampCode)
				require.Equal(t, ledgerItems[i].RunStatus, statusRows[i].RunStatus)
			}
		})
	}
}

// TestOnlyOpenIsStrictSubset 验证维修状态页特有的"仅看未闭环故障"是台账集合的严格子集,
// 且该子集恰好等于"存在未闭环故障"的路灯, 与故障模块的 OpenStatuses 口径一致。
func TestOnlyOpenIsStrictSubset(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	openLamp := h.createLamp(t, "LD-I-101", "解放路")
	h.createLamp(t, "LD-I-102", "解放路")
	h.createFault(t, openLamp.ID, "灯不亮")

	_, allTotal, _, err := h.status.Lamps(ctx, status.LampQuery{RoadName: "解放路"})
	require.NoError(t, err)
	require.Equal(t, int64(2), allTotal)

	openRows, openTotal, _, err := h.status.Lamps(ctx, status.LampQuery{RoadName: "解放路", OnlyOpen: true})
	require.NoError(t, err)
	require.Equal(t, int64(1), openTotal)
	require.Len(t, openRows, 1)
	require.Equal(t, openLamp.Code, openRows[0].LampCode)
	require.Equal(t, int64(1), openRows[0].OpenFaults)
}
