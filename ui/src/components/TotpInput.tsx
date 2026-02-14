import React from "react";
import { Form } from "react-bootstrap";
import { TranslationFunc, withTranslation } from "./withTranslation";

interface State {
}

interface Props {
  t: TranslationFunc;
  required?: boolean | undefined;
  disabled?: boolean | undefined;
  invalid?: boolean | undefined;
  hidden?: boolean | undefined;
  value: string;
  onChange: (value: string) => void;
  onComplete?: (value: string) => void;
}

class TotpInput extends React.Component<Props, State> {
  private inputRefs: React.RefObject<HTMLInputElement | null>[];

  constructor(props: Props) {
    super(props);
    this.state = {
    };
    // Create refs for all 6 input fields
    this.inputRefs = [...Array(6)].map(() => React.createRef<HTMLInputElement>());
  }

  componentDidMount() {
    if (!this.props.hidden) {
      this.focusFirstInput();
    }
  }

  componentDidUpdate(prevProps: Props) {
    // Auto-focus first input when component becomes visible
    if (prevProps.hidden && !this.props.hidden) {
      this.focusFirstInput();
    }
  }

  focusFirstInput = () => {
    setTimeout(() => {
      this.inputRefs[0].current?.focus();
    }, 0);
  }

  getDigits = (): string[] => {
    const value = this.props.value || "";
    const digits = value.padEnd(6, " ").slice(0, 6).split("");
    return digits;
  }

  handleInputChange = (index: number, e: React.ChangeEvent<any>) => {
    const newValue = e.target.value;
    
    // Only allow numeric input
    if (newValue && !/^\d$/.test(newValue)) {
      return;
    }

    const digits = this.getDigits();
    digits[index] = newValue || " ";
    
    const updatedValue = digits.join("").trimEnd();
    this.props.onChange(updatedValue);

    // Auto-focus next input if a digit was entered
    if (newValue && index < 5) {
      this.inputRefs[index + 1].current?.focus();
    }

    // Fire onComplete callback if all 6 digits are entered
    if (updatedValue.length === 6 && this.props.onComplete) {
      this.props.onComplete(updatedValue);
    }
  }

  handleKeyDown = (index: number, e: React.KeyboardEvent<any>) => {
    const input = e.currentTarget;
    
    // Handle backspace
    if (e.key === "Backspace") {
      const digits = this.getDigits();
      
      if (!input.value || input.value === " ") {
        // If current field is empty, move to previous and clear it
        if (index > 0) {
          e.preventDefault();
          digits[index - 1] = " ";
          const updatedValue = digits.join("").trimEnd();
          this.props.onChange(updatedValue);
          this.inputRefs[index - 1].current?.focus();
        }
      } else {
        // Clear current field
        digits[index] = " ";
        const updatedValue = digits.join("").trimEnd();
        this.props.onChange(updatedValue);
      }
    }
    
    // Handle arrow keys
    if (e.key === "ArrowLeft" && index > 0) {
      e.preventDefault();
      this.inputRefs[index - 1].current?.focus();
    }
    
    if (e.key === "ArrowRight" && index < 5) {
      e.preventDefault();
      this.inputRefs[index + 1].current?.focus();
    }
  }

  handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pastedData = e.clipboardData.getData("text");
    
    // Extract only digits from pasted content
    const digits = pastedData.replace(/\D/g, "").slice(0, 6);
    
    if (digits) {
      this.props.onChange(digits);
      
      // Focus the next empty field or the last field
      const nextIndex = Math.min(digits.length, 5);
      this.inputRefs[nextIndex].current?.focus();

      // Fire onComplete callback if all 6 digits are pasted
      if (digits.length === 6 && this.props.onComplete) {
        this.props.onComplete(digits);
      }
    }
  }

  handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
    e.target.select();
  }

  render() {
    const digits = this.getDigits();

    return (
      <div className="d-flex gap-2" style={{ maxWidth: "400px" }}>
        {digits.map((digit, index) => (
          <Form.Control
            key={index}
            ref={this.inputRefs[index]}
            type="text"
            inputMode="numeric"
            maxLength={1}
            value={digit.trim()}
            onChange={(e) => this.handleInputChange(index, e)}
            onKeyDown={(e) => this.handleKeyDown(index, e)}
            onPaste={this.handlePaste}
            onFocus={this.handleFocus}
            disabled={this.props.disabled}
            required={this.props.required}
            isInvalid={this.props.invalid}
            style={{
              flex: 1,
              minWidth: "2.5rem",
              maxWidth: "4rem",
              aspectRatio: "1",
              textAlign: "center",
              fontSize: "1.5rem",
              fontWeight: "bold",
              padding: "0.5rem",
              backgroundImage: "none"
            }}
            className="totp-input"
          />
        ))}
      </div>
    );
  }
}

export default withTranslation(TotpInput as any);
