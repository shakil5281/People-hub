import * as React from "react";
import { SeparationFormData } from "../data/separation-data";
interface SeparationFormProps {
    initialData?: Partial<SeparationFormData>;
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    separationId?: string;
}
export declare function SeparationForm({ initialData, onSuccess, onCancel, isEditing, separationId }: SeparationFormProps): React.JSX.Element;
export {};
