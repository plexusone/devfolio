// Package dashboard provides dashforge-compatible dashboard export.
package dashboard

import "encoding/json"

// Dashboard is the top-level container for a dashboard definition.
// Compatible with dashforge dashboardir.Dashboard.
type Dashboard struct {
	Schema      string       `json:"$schema,omitempty"`
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version,omitempty"`
	Layout      Layout       `json:"layout"`
	DataSources []DataSource `json:"dataSources"`
	Widgets     []Widget     `json:"widgets"`
	Theme       *Theme       `json:"theme,omitempty"`
}

// Layout defines the dashboard layout system.
type Layout struct {
	Type      string `json:"type"`
	Columns   int    `json:"columns,omitempty"`
	RowHeight int    `json:"rowHeight,omitempty"`
	Gap       int    `json:"gap,omitempty"`
	Padding   int    `json:"padding,omitempty"`
}

// Theme defines visual styling.
type Theme struct {
	Mode            string `json:"mode,omitempty"`
	PrimaryColor    string `json:"primaryColor,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
}

// DataSource defines where dashboard data comes from.
type DataSource struct {
	ID          string          `json:"id"`
	Name        string          `json:"name,omitempty"`
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data,omitempty"`
	DerivedFrom string          `json:"derivedFrom,omitempty"`
	Transform   []Transform     `json:"transform,omitempty"`
}

// Transform applies transformations to data.
type Transform struct {
	Type   string `json:"type"`
	Field  string `json:"field,omitempty"`
	Expr   string `json:"expr,omitempty"`
	As     string `json:"as,omitempty"`
	Filter string `json:"filter,omitempty"`
}

// Widget is a visual component on the dashboard.
type Widget struct {
	ID           string          `json:"id"`
	Title        string          `json:"title,omitempty"`
	Description  string          `json:"description,omitempty"`
	Type         string          `json:"type"`
	Position     Position        `json:"position"`
	DataSourceID string          `json:"dataSourceId,omitempty"`
	Config       json.RawMessage `json:"config"`
}

// Position defines widget placement in the grid layout.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// MetricConfig configures a single-value metric widget.
type MetricConfig struct {
	ValueField    string         `json:"valueField"`
	LabelField    string         `json:"labelField,omitempty"`
	Format        string         `json:"format,omitempty"`
	FormatOptions *FormatOptions `json:"formatOptions,omitempty"`
	Icon          string         `json:"icon,omitempty"`
}

// FormatOptions provides format-specific settings.
type FormatOptions struct {
	Decimals int    `json:"decimals,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Suffix   string `json:"suffix,omitempty"`
}

// TableConfig configures a table widget.
type TableConfig struct {
	Columns  []TableColumn `json:"columns"`
	Sortable bool          `json:"sortable,omitempty"`
	Compact  bool          `json:"compact,omitempty"`
}

// TableColumn defines a table column.
type TableColumn struct {
	Field  string `json:"field"`
	Header string `json:"header,omitempty"`
	Width  string `json:"width,omitempty"`
	Align  string `json:"align,omitempty"`
	Format string `json:"format,omitempty"`
}

// ChartIR is the chart intermediate representation.
type ChartIR struct {
	Mark      Mark      `json:"mark"`
	Encodings Encodings `json:"encodings"`
	Data      any       `json:"data,omitempty"`
	Style     *Style    `json:"style,omitempty"`
	Title     string    `json:"title,omitempty"`
}

// Mark defines the chart geometry type.
type Mark struct {
	Type  string     `json:"type"`
	Style *MarkStyle `json:"style,omitempty"`
}

// MarkStyle defines mark-specific styling.
type MarkStyle struct {
	Smooth       bool    `json:"smooth,omitempty"`
	BarWidth     any     `json:"barWidth,omitempty"`
	BorderRadius int     `json:"borderRadius,omitempty"`
	Opacity      float64 `json:"opacity,omitempty"`
}

// Encodings maps data fields to visual properties.
type Encodings struct {
	X        any `json:"x,omitempty"`
	Y        any `json:"y,omitempty"`
	Color    any `json:"color,omitempty"`
	Value    any `json:"value,omitempty"`
	Category any `json:"category,omitempty"`
	Heat     any `json:"heat,omitempty"`
}

// Encoding defines a single field encoding.
type Encoding struct {
	Field string `json:"field"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

// Style defines chart-level styling.
type Style struct {
	Colors     []string `json:"colors,omitempty"`
	Legend     *Legend  `json:"legend,omitempty"`
	XAxis      *Axis    `json:"xAxis,omitempty"`
	YAxis      *Axis    `json:"yAxis,omitempty"`
	Horizontal bool     `json:"horizontal,omitempty"`
	Calendar   *CalendarStyle `json:"calendar,omitempty"`
}

// Legend configures the chart legend.
type Legend struct {
	Show     bool   `json:"show,omitempty"`
	Position string `json:"position,omitempty"`
}

// Axis configures a chart axis.
type Axis struct {
	Show        bool   `json:"show,omitempty"`
	Title       string `json:"title,omitempty"`
	LabelRotate int    `json:"labelRotate,omitempty"`
}

// CalendarStyle configures calendar heatmap styling.
type CalendarStyle struct {
	Range    string   `json:"range,omitempty"`
	CellSize []any    `json:"cellSize,omitempty"`
	Colors   []string `json:"colors,omitempty"`
}

// Widget type constants.
const (
	WidgetTypeChart  = "chart"
	WidgetTypeTable  = "table"
	WidgetTypeMetric = "metric"
	WidgetTypeText   = "text"
)

// Layout type constants.
const (
	LayoutTypeGrid = "grid"
)

// DataSource type constants.
const (
	DataSourceTypeInline  = "inline"
	DataSourceTypeDerived = "derived"
)

// Chart geometry types.
const (
	ChartTypeLine    = "line"
	ChartTypeBar     = "bar"
	ChartTypePie     = "pie"
	ChartTypeHeatmap = "heatmap"
)
