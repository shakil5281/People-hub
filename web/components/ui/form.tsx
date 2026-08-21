"use client"

import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { Controller, ControllerProps, ControllerRenderProps, ControllerFieldState, FieldPath, FieldValues, FormProvider, UseFormStateReturn } from "react-hook-form"
import { cn } from "@/lib/utils"

interface FormProps extends React.ComponentPropsWithoutRef<typeof FormProvider> {}

const Form = FormProvider

type FormFieldContextValue<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>
> = {
  name: TName
}

const FormFieldContext = React.createContext<FormFieldContextValue | null>(null)

interface FormFieldProps<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>
> {
  control: ControllerProps<TFieldValues>["control"]
  name: TName
  render: (props: { field: ControllerRenderProps<TFieldValues, TName>; fieldState: ControllerFieldState; formState: UseFormStateReturn<TFieldValues> }) => React.ReactElement
}

function FormField<TFieldValues extends FieldValues = FieldValues, TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>>({
  control,
  name,
  render,
}: FormFieldProps<TFieldValues, TName>) {
  return (
    <Controller
      control={control}
      name={name}
      render={(props) => render(props)}
    />
  )
}

interface FormItemContextValue {
  id: string
}

const FormItemContext = React.createContext<FormItemContextValue | null>(null)

interface FormItemProps extends React.ComponentPropsWithoutRef<"div"> {}

const FormItem = React.forwardRef<HTMLDivElement, FormItemProps>(({ className, ...props }, ref) => {
  const id = React.useId()
  return (
    <FormItemContext.Provider value={{ id }}>
      <div ref={ref} className={cn("space-y-2", className)} {...props} />
    </FormItemContext.Provider>
  )
})
FormItem.displayName = "FormItem"

interface FormLabelProps extends React.ComponentPropsWithoutRef<"label"> {}

const FormLabel = React.forwardRef<HTMLLabelElement, FormLabelProps>(({ className, ...props }, ref) => {
  const { id } = React.useContext(FormItemContext) ?? {}
  return <label ref={ref} className={cn("", className)} htmlFor={id} {...props} />
})
FormLabel.displayName = "FormLabel"

interface FormControlProps extends React.ComponentPropsWithoutRef<typeof Slot> {}

const FormControl = React.forwardRef<HTMLSlotElement, FormControlProps>(({ className, ...props }, ref) => {
  const { id } = React.useContext(FormItemContext) ?? {}
  return <Slot ref={ref} id={id} className={cn("", className)} {...props} />
})
FormControl.displayName = "FormControl"

interface FormMessageProps extends React.ComponentPropsWithoutRef<"p"> {
  children?: React.ReactNode
}

const FormMessage = React.forwardRef<HTMLParagraphElement, FormMessageProps>(({ className, children, ...props }, ref) => {
  const { id } = React.useContext(FormItemContext) ?? {}
  return (
    <p
      ref={ref}
      id={id ? `${id}-message` : undefined}
      className={cn("text-sm font-medium text-destructive", className)}
      {...props}
    >
      {children}
    </p>
  )
})
FormMessage.displayName = "FormMessage"

export {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
  FormFieldContext,
  FormItemContext,
}