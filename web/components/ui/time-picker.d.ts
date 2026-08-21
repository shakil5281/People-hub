import * as React from "react";
interface TimePickerProps {
    value?: string;
    onChange?: (time: string) => void;
    placeholder?: string;
    className?: string;
    disabled?: boolean;
}
export declare function TimePicker({ value, onChange, placeholder, className, disabled }: TimePickerProps): React.JSX.Element;
export {};
