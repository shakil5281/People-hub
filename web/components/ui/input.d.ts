import * as React from "react";
type InputProps = Omit<React.ComponentProps<"input">, "size"> & {
    size?: "sm" | "md" | "lg";
};
declare function Input({ className, type, size, ...props }: InputProps): React.JSX.Element;
export { Input };
