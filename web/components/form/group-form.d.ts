import * as React from "react";
interface GroupFormProps {
    initialData?: {
        name: string;
    };
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    groupId?: string;
}
export declare function GroupForm({ initialData, onSuccess, onCancel, isEditing, groupId }: GroupFormProps): React.JSX.Element;
export {};
