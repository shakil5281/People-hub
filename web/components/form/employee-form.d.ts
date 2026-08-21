import * as React from "react";
import { EmployeeFormData } from "../data/employee-data";
interface EmployeeFormProps {
    initialData?: Partial<EmployeeFormData>;
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    employeeId?: string;
}
export declare function EmployeeForm({ initialData, onSuccess, onCancel, isEditing, employeeId }: EmployeeFormProps): React.JSX.Element;
export {};
