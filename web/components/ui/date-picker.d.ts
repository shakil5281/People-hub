import * as React from "react";
interface DatePickerProps {
    value?: Date | undefined;
    onChange?: (date: Date | undefined) => void;
    placeholder?: string;
    className?: string;
    disabled?: boolean;
}
export declare function DatePicker({ value, onChange, placeholder, className, disabled }: DatePickerProps): React.JSX.Element;
export {};
