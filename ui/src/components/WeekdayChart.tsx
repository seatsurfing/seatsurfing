import React, { useRef, useEffect } from "react";
import { Card, Col, Row } from "react-bootstrap";
import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
} from "chart.js";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

interface WeekdayChartProps {
  data: number[];
  labels: string[];
  title: string;
  height?: number;
}

const WeekdayChart: React.FC<WeekdayChartProps> = ({
  data,
  labels,
  title,
  height = 200,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const chartRef = useRef<Chart | null>(null);

  useEffect(() => {
    if (!canvasRef.current) return;
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
          y: {
            beginAtZero: true,
            ticks: { stepSize: 1 },
          },
        },
      },
    });
    return () => {
      chartRef.current?.destroy();
    };
  }, [data, labels]);

  return (
    <Row className="mb-4">
      <Col sm="12" xl="8">
        <Card>
          <Card.Body>
            <Card.Title className="mb-3">{title}</Card.Title>
            <div style={{ height: `${height}px` }}>
              <canvas ref={canvasRef} />
            </div>
          </Card.Body>
        </Card>
      </Col>
    </Row>
  );
};

export default WeekdayChart;
