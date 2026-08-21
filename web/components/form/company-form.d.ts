import * as React from "react";
import { CompanyFormData } from "../data/company-data";
interface CompanyFormProps {
    initialData?: Partial<CompanyFormData>;
    onSuccess: () => void;
    onCancel?: () => void;
    isEditing?: boolean;
    companyId?: string;
}
export declare function CompanyForm({ initialData, onSuccess, onCancel, isEditing, companyId }: CompanyFormProps): React.JSX.Element;
export {};
