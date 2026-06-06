package report

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	maroto "github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/varmiguemunoz/sprintos/internal/app"
)

var (
	colorPurple = &props.Color{Red: 124, Green: 58, Blue: 237}
	colorDark   = &props.Color{Red: 30, Green: 30, Blue: 40}
	colorGray   = &props.Color{Red: 110, Green: 110, Blue: 120}
	colorGreen  = &props.Color{Red: 16, Green: 185, Blue: 129}
	colorRed    = &props.Color{Red: 220, Green: 38, Blue: 38}
	colorOrange = &props.Color{Red: 245, Green: 158, Blue: 11}
)

func Generate(data *app.ReportData, destPath string) error {
	cfg := config.NewBuilder().
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(15).
		Build()

	m := maroto.New(cfg)

	m.AddRows(coverSection(data)...)
	m.AddRows(summarySection(data)...)
	m.AddRows(projectHealthSection(data)...)

	if chartBytes, err := renderVelocityChart(data); err == nil {
		m.AddRows(chartSection("Weekly Velocity — Tasks Completed", chartBytes)...)
	}

	m.AddRows(teamSection(data)...)

	if len(data.CriticalHighOpen) > 0 || len(data.OverdueTasks) > 0 {
		m.AddRows(riskSection(data)...)
	}

	doc, err := m.Generate()
	if err != nil {
		return fmt.Errorf("could not generate PDF: %w", err)
	}

	return doc.Save(destPath)
}

func coverSection(data *app.ReportData) []core.Row {
	dateRange := fmt.Sprintf("%s  —  %s",
		data.From.Format("Jan 2, 2006"),
		data.To.Format("Jan 2, 2006"),
	)
	generated := "Generated " + time.Now().Format("January 2, 2006  ·  15:04")

	return []core.Row{
		row.New(12).Add(col.New(12)),
		row.New(20).Add(text.NewCol(12, "SprintOS", props.Text{
			Size:  38,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: colorPurple,
		})),
		row.New(10).Add(text.NewCol(12, "Executive Report", props.Text{
			Size:  16,
			Align: align.Center,
			Color: colorDark,
		})),
		row.New(8).Add(text.NewCol(12, data.OrgName, props.Text{
			Size:  13,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: colorGray,
		})),
		row.New(8).Add(col.New(12)),
		row.New(7).Add(text.NewCol(12, dateRange, props.Text{
			Size:  11,
			Align: align.Center,
			Color: colorPurple,
		})),
		row.New(5).Add(text.NewCol(12, generated, props.Text{
			Size:  9,
			Align: align.Center,
			Color: colorGray,
		})),
		row.New(10).Add(col.New(12)),
		line.NewRow(1),
		row.New(8).Add(col.New(12)),
	}
}

func summarySection(data *app.ReportData) []core.Row {
	return []core.Row{
		sectionTitleRow("Executive Summary"),
		row.New(5).Add(col.New(12)),

		row.New(17).Add(
			kpiValueCol(3, fmt.Sprintf("%d", data.TotalCreated), colorDark),
			kpiValueCol(3, fmt.Sprintf("%d", data.TotalCompleted), colorGreen),
			kpiValueCol(3, fmt.Sprintf("%.1f%%", data.OnTimeRate), colorPurple),
			kpiValueCol(3, fmt.Sprintf("%.1fh", data.TotalHours), colorDark),
		),
		row.New(7).Add(
			kpiLabelCol(3, "Tasks Created"),
			kpiLabelCol(3, "Completed"),
			kpiLabelCol(3, "On-Time Rate"),
			kpiLabelCol(3, "Hours Logged"),
		),
		row.New(5).Add(col.New(12)),

		row.New(14).Add(
			kpiValueCol(4, fmt.Sprintf("%.1f days", data.AvgCycleTimeDays), colorGray),
			kpiValueCol(4, fmt.Sprintf("%d", len(data.OverdueTasks)), colorRed),
			kpiValueCol(4, fmt.Sprintf("%d", len(data.CriticalHighOpen)), colorOrange),
		),
		row.New(7).Add(
			kpiLabelCol(4, "Avg Cycle Time"),
			kpiLabelCol(4, "Overdue Tasks"),
			kpiLabelCol(4, "Critical/High Open"),
		),
		row.New(8).Add(col.New(12)),
		line.NewRow(1),
		row.New(6).Add(col.New(12)),
	}
}

func projectHealthSection(data *app.ReportData) []core.Row {
	if len(data.Projects) == 0 {
		return nil
	}

	rows := []core.Row{
		sectionTitleRow("Project Health"),
		row.New(4).Add(col.New(12)),
		row.New(7).Add(
			headerCell("Project", 4),
			headerCell("Done", 1),
			headerCell("Active", 1),
			headerCell("Backlog", 1),
			headerCell("Overdue", 1),
			headerCell("Hours", 2),
			headerCell("Cycle (d)", 2),
		),
		line.NewRow(1),
	}

	for _, p := range data.Projects {
		rows = append(rows,
			row.New(7).Add(
				dataCell(p.ProjectName, 4, align.Left, colorDark),
				dataCell(strconv.Itoa(p.CompletedTasks), 1, align.Center, colorGreen),
				dataCell(strconv.Itoa(p.InProgressTasks), 1, align.Center, colorDark),
				dataCell(strconv.Itoa(p.BacklogTasks), 1, align.Center, colorGray),
				dataCell(strconv.Itoa(p.OverdueTasks), 1, align.Center, colorRed),
				dataCell(fmt.Sprintf("%.1f", p.TotalHours), 2, align.Center, colorDark),
				dataCell(fmt.Sprintf("%.1f", p.AvgCycleTimeDays), 2, align.Center, colorPurple),
			),
		)
	}

	rows = append(rows,
		line.NewRow(1),
		row.New(8).Add(col.New(12)),
	)

	return rows
}

func chartSection(title string, chartBytes []byte) []core.Row {
	return []core.Row{
		sectionTitleRow(title),
		row.New(4).Add(col.New(12)),
		row.New(65).Add(image.NewFromBytesCol(12, chartBytes, extension.Png)),
		row.New(8).Add(col.New(12)),
		line.NewRow(1),
		row.New(6).Add(col.New(12)),
	}
}

func teamSection(data *app.ReportData) []core.Row {
	if len(data.Members) == 0 {
		return nil
	}

	rows := []core.Row{
		sectionTitleRow("Team Performance"),
		row.New(4).Add(col.New(12)),
		row.New(7).Add(
			headerCell("Member", 6),
			headerCell("Tasks Completed", 3),
			headerCell("Hours Logged", 3),
		),
		line.NewRow(1),
	}

	for _, m := range data.Members {
		rows = append(rows,
			row.New(7).Add(
				dataCell(m.Name, 6, align.Left, colorDark),
				dataCell(strconv.Itoa(m.TasksCompleted), 3, align.Center, colorGreen),
				dataCell(fmt.Sprintf("%.1f", m.HoursLogged), 3, align.Center, colorDark),
			),
		)
	}

	rows = append(rows, line.NewRow(1), row.New(8).Add(col.New(12)))
	return rows
}

func riskSection(data *app.ReportData) []core.Row {
	rows := []core.Row{
		sectionTitleRow("Priority Risk"),
		row.New(4).Add(col.New(12)),
	}

	if len(data.CriticalHighOpen) > 0 {
		rows = append(rows,
			row.New(6).Add(text.NewCol(12, "Critical & High Priority — Open Tasks", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Color: colorOrange,
			})),
			row.New(3).Add(col.New(12)),
			row.New(7).Add(
				headerCell("Task", 5),
				headerCell("Project", 3),
				headerCell("Priority", 2),
				headerCell("Age (days)", 2),
			),
			line.NewRow(1),
		)
		for _, t := range data.CriticalHighOpen {
			priColor := colorOrange
			if t.Priority == "critical" {
				priColor = colorRed
			}
			rows = append(rows,
				row.New(7).Add(
					dataCell(t.Title, 5, align.Left, colorDark),
					dataCell(t.ProjectName, 3, align.Left, colorGray),
					dataCell(t.Priority, 2, align.Center, priColor),
					dataCell(strconv.Itoa(t.AgeDays), 2, align.Center, colorDark),
				),
			)
		}
		rows = append(rows, line.NewRow(1), row.New(6).Add(col.New(12)))
	}

	if len(data.OverdueTasks) > 0 {
		rows = append(rows,
			row.New(6).Add(text.NewCol(12, "Overdue Tasks", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Color: colorRed,
			})),
			row.New(3).Add(col.New(12)),
			row.New(7).Add(
				headerCell("Task", 5),
				headerCell("Project", 3),
				headerCell("Priority", 2),
				headerCell("Days Overdue", 2),
			),
			line.NewRow(1),
		)
		for _, t := range data.OverdueTasks {
			rows = append(rows,
				row.New(7).Add(
					dataCell(t.Title, 5, align.Left, colorDark),
					dataCell(t.ProjectName, 3, align.Left, colorGray),
					dataCell(t.Priority, 2, align.Center, colorOrange),
					dataCell(strconv.Itoa(t.AgeDays), 2, align.Center, colorRed),
				),
			)
		}
		rows = append(rows, line.NewRow(1))
	}

	return rows
}

func renderVelocityChart(data *app.ReportData) ([]byte, error) {
	if len(data.WeeklyVelocity) == 0 {
		return nil, fmt.Errorf("no velocity data")
	}

	purple := drawing.Color{R: 124, G: 58, B: 237, A: 255}

	bars := make([]chart.Value, len(data.WeeklyVelocity))
	for i, v := range data.WeeklyVelocity {
		bars[i] = chart.Value{
			Value: float64(v.Completed),
			Label: v.WeekLabel,
			Style: chart.Style{
				FillColor:   purple,
				StrokeColor: purple,
			},
		}
	}

	bc := chart.BarChart{
		Background: chart.Style{
			FillColor: drawing.ColorWhite,
			Padding: chart.Box{
				Top:    30,
				Left:   20,
				Right:  20,
				Bottom: 10,
			},
		},
		Height:   280,
		BarWidth: 40,
		XAxis: chart.Style{
			FontSize:  9,
			FontColor: drawing.Color{R: 110, G: 110, B: 120, A: 255},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize:  9,
				FontColor: drawing.Color{R: 110, G: 110, B: 120, A: 255},
			},
		},
		Bars: bars,
	}

	buf := &bytes.Buffer{}
	if err := bc.Render(chart.PNG, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sectionTitleRow(title string) core.Row {
	return row.New(10).Add(text.NewCol(12, title, props.Text{
		Size:  13,
		Style: fontstyle.Bold,
		Color: colorPurple,
	}))
}

func headerCell(label string, size int) core.Col {
	return text.NewCol(size, label, props.Text{
		Size:  8,
		Style: fontstyle.Bold,
		Color: colorDark,
	})
}

func dataCell(value string, size int, a align.Type, color *props.Color) core.Col {
	return text.NewCol(size, value, props.Text{
		Size:  8,
		Align: a,
		Color: color,
		Top:   1,
	})
}

func kpiValueCol(size int, value string, color *props.Color) core.Col {
	return text.NewCol(size, value, props.Text{
		Size:  22,
		Style: fontstyle.Bold,
		Align: align.Center,
		Color: color,
	})
}

func kpiLabelCol(size int, label string) core.Col {
	return text.NewCol(size, label, props.Text{
		Size:  8,
		Align: align.Center,
		Color: colorGray,
	})
}
