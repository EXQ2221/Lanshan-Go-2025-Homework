package geo

import (
	"fmt"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// GeoInfo 地理位置信息
type GeoInfo struct {
	Country  string
	Region   string // 省/州
	Province string
	City     string
	ISP      string
}

// IPGeoLocator IP地理位置查询器
type IPGeoLocator struct {
	searcher *xdb.Searcher
}

// NewIPGeoLocator 创建 IP 定位器（使用 VectorIndex 缓存策略）
func NewIPGeoLocator(dbPath string) (*IPGeoLocator, error) {
	// 加载 VectorIndex 索引
	vIndex, err := xdb.LoadVectorIndexFromFile(dbPath)
	if err != nil {
		return nil, fmt.Errorf("load vector index failed: %w", err)
	}

	// 创建 searcher
	searcher, err := xdb.NewWithVectorIndex(xdb.IPv4, dbPath, vIndex)
	if err != nil {
		return nil, fmt.Errorf("create searcher failed: %w", err)
	}

	return &IPGeoLocator{searcher: searcher}, nil
}

// Close 关闭查询器
func (l *IPGeoLocator) Close() {
	if l.searcher != nil {
		l.searcher.Close()
	}
}

// Search 查询 IP 地理位置
func (l *IPGeoLocator) Search(ip string) (*GeoInfo, error) {
	region, err := l.searcher.Search(ip)
	if err != nil {
		return nil, err
	}

	// ip2region 返回格式: 国家|区域|省份|城市|ISP
	parts := strings.Split(region, "|")
	info := &GeoInfo{
		Country:  safeGet(parts, 0),
		Region:   safeGet(parts, 1),
		Province: safeGet(parts, 2),
		City:     safeGet(parts, 3),
		ISP:      safeGet(parts, 4),
	}

	return info, nil
}

// GetCityKey 获取城市标识（用于比对是否异地）
// 返回格式: 省份|城市，如果都为空则返回 "unknown"
func (g *GeoInfo) GetCityKey() string {
	if g.Province == "" && g.City == "" {
		return "unknown"
	}
	return g.Province + "|" + g.City
}

func safeGet(arr []string, idx int) string {
	if idx < len(arr) {
		return arr[idx]
	}
	return ""
}
