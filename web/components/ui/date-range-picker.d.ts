import * as React from "react";
import { type DateRange } from "date-fns";
interface DateRangePickerProps {
    value?: DateRange;
    onChange?: (range: DateRange | undefined) => void;
    placeholder?: string;
    className?: string;
    disabled?: boolean;
}
export declare function DateRangePicker({ value, onChange, placeholder, className, disabled }: DateRangePickerProps): React.JSX.Element;
export {};
