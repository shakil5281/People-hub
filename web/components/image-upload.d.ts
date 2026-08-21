import * as React from "react";
interface ImageUploadProps {
    value?: string;
    onChange: (url: string) => void;
    label?: string;
    accept?: string;
}
export declare function ImageUpload({ value, onChange, label, accept }: ImageUploadProps): React.JSX.Element;
export {};
