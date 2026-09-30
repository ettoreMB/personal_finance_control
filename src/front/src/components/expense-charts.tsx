"use client"

import { Bar, BarChart, Cell, Label, Pie, PieChart, XAxis, YAxis } from "recharts"

import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export type ExpenseSlice = {
  category: string
  cents: number
}

const SLICE_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
  "oklch(0.623 0.214 259.815)",
  "oklch(0.627 0.265 303.9)",
  "oklch(0.645 0.246 16.439)",
  "oklch(0.696 0.17 162.48)",
  "oklch(0.72 0.16 130)",
]

function formatAxisReais(reais: number): string {
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
    maximumFractionDigits: 0,
  }).format(reais)
}

export function ExpenseCharts({
  slices,
  periodLabel,
  formatBRL,
}: {
  slices: ExpenseSlice[]
  periodLabel: string
  formatBRL: (cents: number) => string
}) {
  const rows = slices
    .filter((slice) => slice.cents > 0)
    .map((slice, index) => ({
      category: slice.category,
      reais: slice.cents / 100,
      cents: slice.cents,
      fill: SLICE_COLORS[index % SLICE_COLORS.length],
    }))

  if (rows.length === 0) {
    return null
  }

  const totalCents = rows.reduce((sum, row) => sum + row.cents, 0)
  const config: ChartConfig = {
    reais: { label: "Gastos" },
  }
  for (const row of rows) {
    config[row.category] = { label: row.category }
  }

  const barHeight = Math.max(240, rows.length * 36)

  return (
    <div className="grid grid-cols-1 gap-4 @4xl/main:grid-cols-5">
      <Card className="@4xl/main:col-span-3">
        <CardHeader>
          <CardTitle>Gastos por categoria</CardTitle>
          <CardDescription>{periodLabel}</CardDescription>
        </CardHeader>
        <CardContent>
          <ChartContainer
            config={config}
            className="aspect-auto w-full"
            style={{ height: barHeight }}
            initialDimension={{ width: 480, height: barHeight }}
          >
            <BarChart
              accessibilityLayer
              data={rows}
              layout="vertical"
              margin={{ left: 8, right: 12 }}
            >
              <XAxis
                type="number"
                dataKey="reais"
                tickLine={false}
                axisLine={false}
                tickFormatter={formatAxisReais}
              />
              <YAxis
                type="category"
                dataKey="category"
                tickLine={false}
                axisLine={false}
                width={96}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    labelKey="category"
                    formatter={(value, _name, _item, _index, payload) => {
                      const category =
                        payload &&
                        typeof payload === "object" &&
                        "category" in payload
                          ? String(payload.category)
                          : "Gastos"
                      return (
                        <div className="flex w-full items-center justify-between gap-4">
                          <span className="text-muted-foreground">
                            {category}
                          </span>
                          <span className="font-mono font-medium tabular-nums">
                            {formatBRL(Math.round(Number(value) * 100))}
                          </span>
                        </div>
                      )
                    }}
                  />
                }
              />
              <Bar dataKey="reais" radius={6}>
                {rows.map((row) => (
                  <Cell key={row.category} fill={row.fill} />
                ))}
              </Bar>
            </BarChart>
          </ChartContainer>
        </CardContent>
      </Card>
      <Card className="@4xl/main:col-span-2">
        <CardHeader>
          <CardTitle>Participação</CardTitle>
          <CardDescription>Quanto cada categoria pesa nos gastos</CardDescription>
        </CardHeader>
        <CardContent>
          <ChartContainer
            config={config}
            className="mx-auto aspect-square max-h-[320px]"
            initialDimension={{ width: 320, height: 320 }}
          >
            <PieChart>
              <ChartTooltip
                content={
                  <ChartTooltipContent
                    nameKey="category"
                    hideLabel
                    formatter={(value, name, _item, _index, payload) => {
                      const category =
                        payload &&
                        typeof payload === "object" &&
                        "category" in payload
                          ? String(payload.category)
                          : String(name)
                      return (
                        <div className="flex w-full items-center justify-between gap-4">
                          <span className="text-muted-foreground">
                            {category}
                          </span>
                          <span className="font-mono font-medium tabular-nums">
                            {formatBRL(Math.round(Number(value) * 100))}
                          </span>
                        </div>
                      )
                    }}
                  />
                }
              />
              <Pie
                data={rows}
                dataKey="reais"
                nameKey="category"
                innerRadius={68}
                outerRadius={100}
                strokeWidth={2}
                paddingAngle={2}
              >
                {rows.map((row) => (
                  <Cell key={row.category} fill={row.fill} />
                ))}
                <Label
                  content={({ viewBox }) => {
                    if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                      return (
                        <text
                          x={viewBox.cx}
                          y={viewBox.cy}
                          textAnchor="middle"
                          dominantBaseline="middle"
                        >
                          <tspan
                            x={viewBox.cx}
                            y={viewBox.cy}
                            className="fill-foreground text-base font-semibold"
                          >
                            {formatBRL(totalCents)}
                          </tspan>
                          <tspan
                            x={viewBox.cx}
                            y={(viewBox.cy ?? 0) + 18}
                            className="fill-muted-foreground text-xs"
                          >
                            em gastos
                          </tspan>
                        </text>
                      )
                    }
                    return null
                  }}
                />
              </Pie>
              <ChartLegend
                content={<ChartLegendContent nameKey="category" />}
              />
            </PieChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  )
}
