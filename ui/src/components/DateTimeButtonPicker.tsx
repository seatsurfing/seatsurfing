import React, { useRef, useState } from "react";
import { Button, Overlay, Popover, Row, Col } from "react-bootstrap";
import Calendar from "react-calendar";
// We'll implement a simple digital time selector instead of the analog clock.
import "react-calendar/dist/Calendar.css";
import "react-clock/dist/Clock.css";
import Formatting from "@/util/Formatting";

interface Props {
  value: Date;
  onChange: (d: Date) => void;
  disabled?: boolean;
  dailyOnly?: boolean; // when true, show only date button
  showTodayButton?: boolean; // when false, hide Today button (e.g., for end/leave)
}

function zeroPad(v: number) {
  return v.toString().padStart(2, "0");
}

const DateTimeButtonPicker: React.FC<Props> = ({
  value,
  onChange,
  disabled,
  dailyOnly,
  showTodayButton,
}) => {
  const dateBtnRef = useRef<HTMLButtonElement | null>(null);
  const timeBtnRef = useRef<HTMLButtonElement | null>(null);
  const [showDate, setShowDate] = useState(false);
  const [showTime, setShowTime] = useState(false);
  // Default to calendar view
  const [dateViewMode, setDateViewMode] = useState<"calendar" | "select">(
    "calendar",
  );

  const showDatePopover = (next = !showDate) => {
    setShowDate(next);
    if (next) setShowTime(false);
  };
  const showTimePopover = (next = !showTime) => {
    setShowTime(next);
    if (next) setShowDate(false);
  };

  const onDateSelect = (d: Date) => {
    const nd = new Date(value);
    nd.setFullYear(d.getFullYear(), d.getMonth(), d.getDate());
    onChange(nd);
    setShowDate(false);
  };

  const onTimeSelect = (d: Date | null) => {
    if (!d) return;
    const nd = new Date(value);
    nd.setHours(d.getHours(), d.getMinutes(), 0, 0);
    onChange(nd);
    setShowTime(false);
  };

  const onTodayClick = () => {
    if (dailyOnly) {
      const nd = new Date();
      nd.setHours(0, 0, 0, 0);
      onChange(nd);
    } else {
      // Always move to the next full minute so the current minute is not selectable.
      const nd = new Date();
      nd.setSeconds(0, 0);
      nd.setMinutes(nd.getMinutes() + 1);
      onChange(nd);
    }
    setShowDate(false);
    setShowTime(false);
  };

  const dateLabel = (() => {
    const day = zeroPad(value.getDate());
    const month = zeroPad(value.getMonth() + 1);
    const year = value.getFullYear();
    return `${day}.${month}.${year}`;
  })();

  const today = new Date();
  const now = new Date();
  // Minimum allowed booking time is the next full minute. This ensures the current
  // minute is not selectable (e.g. if it's 12:01:30 you can only pick 12:02 or later).
  const minAllowed = new Date();
  minAllowed.setSeconds(0, 0);
  minAllowed.setMinutes(minAllowed.getMinutes() + 1);

  const startOfToday = new Date();
  startOfToday.setHours(0, 0, 0, 0);

  function isPastDate(y: number, m: number, d: number) {
    const cand = new Date(y, m, d);
    cand.setHours(0, 0, 0, 0);
    return cand < startOfToday;
  }

  const timeLabel = (() => {
    const hh = zeroPad(value.getHours());
    const mm = zeroPad(value.getMinutes());
    return `${hh}:${mm}`;
  })();

  return (
    <div className="d-flex align-items-center" style={{ gap: 8 }}>
      <div>
        <Button
          ref={dateBtnRef}
          variant="outline-secondary"
          onClick={() => showDatePopover()}
          disabled={disabled}
          style={{ minWidth: 120, textAlign: "left", color: '#000' }}
        >
          {dateLabel}
        </Button>
        <Overlay
          show={showDate}
          target={dateBtnRef.current}
          placement="top"
          rootClose={true}
          onHide={() => setShowDate(false)}
        >
          <Popover id="popover-date">
            <Popover.Body>
              <div className="d-flex gap-2 mb-2">
                <Button
                  size="sm"
                  variant={dateViewMode === "calendar" ? "primary" : "outline-secondary"}
                  onClick={() => setDateViewMode("calendar")}
                >
                  Kalender
                </Button>
                <Button
                  size="sm"
                  variant={dateViewMode === "select" ? "primary" : "outline-secondary"}
                  onClick={() => setDateViewMode("select")}
                >
                  DD.MM.YYYY
                </Button>
              </div>
                {dateViewMode === "calendar" ? (
                <Calendar
                  minDate={today}
                  onClickDay={(d: Date) => onDateSelect(d)}
                  value={value}
                  // hide double navigation (<< and >>)
                  prev2Label={null}
                  next2Label={null}
                />
              ) : (
                <div>
                  <Row className="g-2">
                    <Col xs={4}>
                      <select
                        className="form-select"
                        value={value.getDate()}
                        onChange={(e) => {
                          const day = parseInt(e.target.value, 10);
                          const nd = new Date(value);
                          nd.setDate(day);
                          onChange(nd);
                        }}
                      >
                        {Array.from({ length: 31 }, (_, i) => i + 1).map((d) => {
                          const disabled = isPastDate(value.getFullYear(), value.getMonth(), d);
                          return (
                            <option key={d} value={d} disabled={disabled}>{d.toString().padStart(2,'0')}</option>
                          );
                        })}
                      </select>
                    </Col>
                    <Col xs={4}>
                      <select
                        className="form-select"
                        value={value.getMonth() + 1}
                        onChange={(e) => {
                          const m = parseInt(e.target.value, 10) - 1;
                          const nd = new Date(value);
                          nd.setMonth(m);
                          onChange(nd);
                        }}
                      >
                        {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => {
                          const disabled = isPastDate(value.getFullYear(), m - 1, value.getDate());
                          return (
                            <option key={m} value={m} disabled={disabled}>{m.toString().padStart(2,'0')}</option>
                          );
                        })}
                      </select>
                    </Col>
                    <Col xs={4}>
                      <select
                        className="form-select"
                        style={{ minWidth: '90px' }}
                        value={value.getFullYear()}
                        onChange={(e) => {
                          const y = parseInt(e.target.value, 10);
                          const nd = new Date(value);
                          nd.setFullYear(y);
                          onChange(nd);
                        }}
                      >
                        {Array.from({ length: 21 }, (_, i) => value.getFullYear() - 10 + i).map((y) => {
                          const disabled = isPastDate(y, value.getMonth(), value.getDate());
                          return (
                            <option key={y} value={y} disabled={disabled}>{y}</option>
                          );
                        })}
                      </select>
                    </Col>
                  </Row>
                </div>
              )}
            </Popover.Body>
          </Popover>
        </Overlay>
      </div>

      {!dailyOnly && (
        <div>
          <Button
            ref={timeBtnRef}
            variant="outline-secondary"
            onClick={() => showTimePopover()}
            disabled={disabled}
            style={{ minWidth: 80, color: '#000' }}
          >
            {timeLabel}
          </Button>
          <Overlay
            show={showTime}
            target={timeBtnRef.current}
            placement="top"
            rootClose={true}
            onHide={() => setShowTime(false)}
          >
            <Popover id="popover-time">
              <Popover.Body>
                <Row className="g-2 align-items-center">
                    <Col xs={6}>
                    <select
                      className="form-select"
                      value={value.getHours()}
                      onChange={(e) => {
                        const hh = parseInt(e.target.value, 10);
                        const nd = new Date(value);
                        nd.setHours(hh);
                        onChange(nd);
                      }}
                    >
                      {Array.from({ length: 24 }, (_, i) => i).map((h) => {
                        const cand = new Date(value);
                        cand.setHours(h, 0, 0, 0);
                        const disabled = cand < minAllowed;
                        return (
                          <option key={h} value={h} disabled={disabled}>{h.toString().padStart(2,'0')}</option>
                        );
                      })}
                    </select>
                  </Col>
                  <Col xs={6}>
                    <select
                      className="form-select"
                      value={value.getMinutes()}
                      onChange={(e) => {
                        const mm = parseInt(e.target.value, 10);
                        const nd = new Date(value);
                        nd.setMinutes(mm);
                        onChange(nd);
                      }}
                    >
                      {Array.from({ length: 60 }, (_, i) => i).map((m) => {
                        const cand = new Date(value);
                        cand.setHours(value.getHours(), m, 0, 0);
                        const disabled = cand < minAllowed;
                        return (
                          <option key={m} value={m} disabled={disabled}>{m.toString().padStart(2,'0')}</option>
                        );
                      })}
                    </select>
                  </Col>
                </Row>
              </Popover.Body>
            </Popover>
          </Overlay>
        </div>
      )}
      {/* Today button - sets date/time to now (or start of day when dailyOnly) */}
      {showTodayButton !== false && (
        <div>
          <Button
            size="sm"
            variant="outline-secondary"
            onClick={onTodayClick}
            disabled={disabled}
          >
            {Formatting.t("today")}
          </Button>
        </div>
      )}
    </div>
  );
};

export default DateTimeButtonPicker;
