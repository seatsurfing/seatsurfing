import React, { useRef, useEffect, useState } from "react";
import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
} from "chart.js";
import Theme from "@/util/Theme";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

interface WeekdayChartProps {
  data: number[];
  labels: string[];
  height?: number;
}

const WeekdayChart: React.FC<WeekdayChartProps> = ({
  data,
  labels,
  height = 200,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const chartRef = useRef<Chart | null>(null);
  const [theme, setTheme] = useState(Theme.getEffective());

  useEffect(() => {
    const onThemeChange = () => setTheme(Theme.getEffective());
    window.addEventListener(Theme.CHANGE_EVENT, onThemeChange);
    return () => window.removeEventListener(Theme.CHANGE_EVENT, onThemeChange);
  }, []);

  useEffect(() => {
    if (!canvasRef.current) return;
    const styles = getComputedStyle(document.documentElement);
    const textColor = styles.getPropertyValue("--bs-secondary-color");
    const gridColor = styles.getPropertyValue("--bs-border-color-translucent");
    if (chartRef.current) {
      chartRef.current.destroy();
    }
    chartRef.current = new Chart(canvasRef.current, {
      type: "bar",
      data: {
        labels,
        datasets: [
          {
            data,
            backgroundColor: "#0d6efd",
            borderRadius: 3,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        events: [],
        plugins: {
          legend: { display: false },
          tooltip: { enabled: false },
        },
        scales: {
          x: {
            ticks: { color: textColor },
            grid: { color: gridColor },
          },
          y: {
            beginAtZero: true,
            ticks: { stepSize: 1, color: textColor },
            grid: { color: gridColor },
          },
        },
      },
    });
    return () => {
      chartRef.current?.destroy();
    };
  }, [data, labels, theme]);

  return (
    <div style={{ height: `${height}px` }}>
      <canvas ref={canvasRef} />
    </div>
  );
};

export default WeekdayChart;
