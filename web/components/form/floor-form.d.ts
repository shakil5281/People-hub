import * as React from "react";
interface FloorFormProps {
    initialData?: {
        name: string;
    };
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    floorId?: string;
}
export declare function FloorForm({ initialData, onSuccess, onCancel, isEditing, floorId }: FloorFormProps): React.JSX.Element;
export {};
