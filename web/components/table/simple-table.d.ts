import * as React from "react";
interface Column<T> {
    header: string;
    accessor: keyof T | ((item: T) => string | number);
    className?: string;
    cell?: (item: T) => React.ReactNode;
}
interface SimpleTableProps<T> {
    title: string;
    description?: string;
    data: T[];
    columns: Column<T>[];
    searchKey?: keyof T;
    pageSize?: number;
}
export declare function SimpleTable<T extends Record<string, unknown>>({ title, description, data, columns, searchKey, pageSize, }: SimpleTableProps<T>): React.JSX.Element;
export {};
