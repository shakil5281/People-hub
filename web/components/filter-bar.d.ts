import * as React from "react";
export interface FilterDef {
    key: string;
    label: string;
    type: "select" | "text" | "number" | "date" | "datepicker" | "daterange";
    options?: {
        value: string;
        label: string;
    }[];
    placeholder?: string;
    disabled?: boolean;
    dateRangeKeys?: {
        start: string;
        end: string;
    };
}
interface FilterBarProps {
    filters: FilterDef[];
    values: Record<string, string>;
    onChange: (key: string, value: string) => void;
    onApply: () => void;
    onReset: () => void;
    submitting: boolean;
}
export declare function FilterBar({ filters, values, onChange, onApply, onReset, submitting }: FilterBarProps): React.JSX.Element;
export {};
