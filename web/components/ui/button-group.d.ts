import { type VariantProps } from "class-variance-authority";
import { Separator } from "@/components/ui/separator";
declare const buttonGroupVariants: (props?: ({
    orientation?: "horizontal" | "vertical" | null | undefined;
} & import("class-variance-authority/types").ClassProp) | undefined) => string;
declare function ButtonGroup({ className, orientation, ...props }: React.ComponentProps<"div"> & VariantProps<typeof buttonGroupVariants>): import("react").JSX.Element;
declare function ButtonGroupText({ className, asChild, ...props }: React.ComponentProps<"div"> & {
    asChild?: boolean;
}): import("react").JSX.Element;
declare function ButtonGroupSeparator({ className, orientation, ...props }: React.ComponentProps<typeof Separator>): import("react").JSX.Element;
export { ButtonGroup, ButtonGroupSeparator, ButtonGroupText, buttonGroupVariants, };
