import * as React from "react";
import { IdCardFormData } from "../data/id-card-data";
interface IdCardFormProps {
    initialData?: Partial<IdCardFormData>;
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    cardId?: string;
}
export declare function IdCardForm({ initialData, onSuccess, onCancel, isEditing, cardId }: IdCardFormProps): React.JSX.Element;
export {};
