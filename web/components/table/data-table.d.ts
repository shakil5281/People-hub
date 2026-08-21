import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
interface DataTableProps<TData extends {
    id: number | string;
}> {
    data: TData[];
    columns: ColumnDef<TData>[];
    enableDnd?: boolean;
    onEdit?: (row: TData) => void;
    onDelete?: (row: TData) => void;
}
export declare function DataTable<TData extends {
    id: number | string;
}>({ data: initialData, columns, enableDnd, onEdit, onDelete, }: DataTableProps<TData>): React.JSX.Element;
export {};
