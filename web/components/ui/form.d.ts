import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { ControllerProps, ControllerRenderProps, ControllerFieldState, FieldPath, FieldValues, UseFormStateReturn } from "react-hook-form";
declare const Form: <TFieldValues extends FieldValues, TContext = any, TTransformedValues = TFieldValues>({ children, watch, getValues, getFieldState, setError, clearErrors, setValue, setValues, trigger, formState, resetField, reset, handleSubmit, unregister, control, register, setFocus, subscribe, }: import("react-hook-form").FormProviderProps<TFieldValues, TContext, TTransformedValues>) => React.JSX.Element;
type FormFieldContextValue<TFieldValues extends FieldValues = FieldValues, TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>> = {
    name: TName;
};
declare const FormFieldContext: React.Context<FormFieldContextValue<FieldValues, string> | null>;
interface FormFieldProps<TFieldValues extends FieldValues = FieldValues, TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>> {
    control: ControllerProps<TFieldValues>["control"];
    name: TName;
    render: (props: {
        field: ControllerRenderProps<TFieldValues, TName>;
        fieldState: ControllerFieldState;
        formState: UseFormStateReturn<TFieldValues>;
    }) => React.ReactElement;
}
declare function FormField<TFieldValues extends FieldValues = FieldValues, TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>>({ control, name, render, }: FormFieldProps<TFieldValues, TName>): React.JSX.Element;
interface FormItemContextValue {
    id: string;
}
declare const FormItemContext: React.Context<FormItemContextValue | null>;
interface FormItemProps extends React.ComponentPropsWithoutRef<"div"> {
}
declare const FormItem: React.ForwardRefExoticComponent<FormItemProps & React.RefAttributes<HTMLDivElement>>;
interface FormLabelProps extends React.ComponentPropsWithoutRef<"label"> {
}
declare const FormLabel: React.ForwardRefExoticComponent<FormLabelProps & React.RefAttributes<HTMLLabelElement>>;
interface FormControlProps extends React.ComponentPropsWithoutRef<typeof Slot> {
}
declare const FormControl: React.ForwardRefExoticComponent<FormControlProps & React.RefAttributes<HTMLSlotElement>>;
interface FormMessageProps extends React.ComponentPropsWithoutRef<"p"> {
    children?: React.ReactNode;
}
declare const FormMessage: React.ForwardRefExoticComponent<FormMessageProps & React.RefAttributes<HTMLParagraphElement>>;
export { Form, FormField, FormItem, FormLabel, FormControl, FormMessage, FormFieldContext, FormItemContext, };
