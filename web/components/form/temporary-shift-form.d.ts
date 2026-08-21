import * as React from "react";
import { TempShiftFormData } from "../data/temporary-shift-data";
interface TempShiftFormProps {
    initialData?: Partial<TempShiftFormData>;
    onSuccess: (data: TempShiftFormData) => void;
    onCancel?: () => void;
    isEditing?: boolean;
    tempShiftId?: number;
}
export declare function TempShiftForm({ initialData, onSuccess, onCancel, isEditing, tempShiftId }: TempShiftFormProps): React.JSX.Element;
export {};
