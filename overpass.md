# Overpass API 使い方メモ

## エンドポイント

```
POST https://overpass-api.de/api/interpreter
Content-Type: application/x-www-form-urlencoded
Body: data=<OverpassQLクエリ>
```

## クエリ（Overpass QL）

横浜市内の道路ネットワークを取得する典型例：

```
[out:json];
(
  way["highway"](35.40,139.55,35.55,139.75);
  >;
);
out body;
```

- `[out:json]` — JSON形式で出力
- `way["highway"](南緯,西経,北緯,東経)` — bounding box内のhighwayタグ付きwayを取得
- `>;` — wayに含まれるnodeも一緒に取得（これがないとノード座標が得られない）
- `out body;` — タグ・座標を含むフル出力

### highway の絞り込み例

全道路を取得すると大量になるため、道路種別で絞るのが現実的：

```
way["highway"~"^(motorway|trunk|primary|secondary|tertiary|residential)$"](bbox);
```

## レスポンスのJSON構造

```json
{
  "elements": [
    {
      "type": "node",
      "id": 123456789,
      "lat": 35.4500,
      "lon": 139.6300,
      "tags": {}
    },
    {
      "type": "way",
      "id": 987654321,
      "nodes": [123456789, 123456790, 123456791],
      "tags": {
        "highway": "primary",
        "name": "国道1号",
        "maxspeed": "60"
      }
    }
  ]
}
```

- **node**: `id`, `lat`, `lon` を持つ。グラフのノード（交差点・道路上の点）
- **way**: `nodes`（node IDの配列）と `tags` を持つ。グラフのエッジ列に相当

## グラフへの変換方法

1. `node` 要素をパース → `map[int64]{lat, lon}` を作る
2. `way` 要素をパース → `nodes` 配列を順番に見て、隣接するnode間にエッジを張る
3. エッジ重み = 隣接ノード間のハーバーサイン距離（緯度経度 → メートル）

```
way.nodes = [A, B, C, D]
→ エッジ: A-B, B-C, C-D（双方向なら逆も）
```

## Go 最小サンプル

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"net/http"
)

type OverpassResponse struct {
	Elements []Element `json:"elements"`
}

type Element struct {
	Type  string            `json:"type"`
	ID    int64             `json:"id"`
	Lat   float64           `json:"lat,omitempty"`
	Lon   float64           `json:"lon,omitempty"`
	Nodes []int64           `json:"nodes,omitempty"`
	Tags  map[string]string `json:"tags,omitempty"`
}

func fetchOSM(query string) (*OverpassResponse, error) {
	body := url.Values{}
	body.Set("data", query)

	resp, err := http.Post(
		"https://overpass-api.de/api/interpreter",
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(body.Encode()),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result OverpassResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func main() {
	query := `[out:json];
(
  way["highway"~"^(primary|secondary|tertiary|residential)$"](35.40,139.55,35.55,139.75);
  >;
);
out body;`

	result, err := fetchOSM(query)
	if err != nil {
		panic(err)
	}

	nodes := make(map[int64][2]float64) // id -> {lat, lon}
	for _, e := range result.Elements {
		if e.Type == "node" {
			nodes[e.ID] = [2]float64{e.Lat, e.Lon}
		}
	}

	// wayからエッジを構築
	for _, e := range result.Elements {
		if e.Type == "way" {
			for i := 0; i < len(e.Nodes)-1; i++ {
				a, b := e.Nodes[i], e.Nodes[i+1]
				fmt.Printf("edge: %d -> %d\n", a, b)
				_ = nodes[a] // {lat, lon} でハーバーサイン距離を計算してエッジ重みに
			}
		}
	}
}
```

## 注意点

- **レート制限**: 過負荷時は429が返る。取得は1回だけ行ってJSONをローカルに保存し、以降はそれを読み込む
- **bounding box のサイズ**: 広すぎると数十万ノードになりBellman-Fordが実用不可になる。横浜市の一区（例: 港北区）程度に絞るのが妥当
- **一方通行**: `oneway=yes` タグが付いたwayは有向エッジにする必要がある（今回は無視して双方向としても可）
