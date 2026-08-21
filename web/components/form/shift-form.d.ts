import * as React from "react";
import { ShiftFormData } from "../data/shift-data";
interface ShiftFormProps {
    initialData?: Partial<ShiftFormData>;
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    shiftId?: string;
}
export declare function ShiftForm({ initialData, onSuccess, onCancel, isEditing, shiftId }: ShiftFormProps): React.JSX.Element;
export {};
