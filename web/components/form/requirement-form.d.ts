import * as React from "react";
import { RequirementFormData } from "../data/requirement-data";
interface RequirementFormProps {
    initialData?: Partial<RequirementFormData>;
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    requirementId?: string;
}
export declare function RequirementForm({ initialData, onSuccess, onCancel, isEditing, requirementId }: RequirementFormProps): React.JSX.Element;
export {};
