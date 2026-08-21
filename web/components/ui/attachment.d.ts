import * as React from "react";
import { type VariantProps } from "class-variance-authority";
import { Button } from "@/components/ui/button";
declare const attachmentVariants: (props?: ({
    size?: "default" | "xs" | "sm" | null | undefined;
    orientation?: "horizontal" | "vertical" | null | undefined;
} & import("class-variance-authority/types").ClassProp) | undefined) => string;
declare function Attachment({ className, state, size, orientation, ...props }: React.ComponentProps<"div"> & VariantProps<typeof attachmentVariants> & {
    state?: "idle" | "uploading" | "processing" | "error" | "done";
}): React.JSX.Element;
declare const attachmentMediaVariants: (props?: ({
    variant?: "icon" | "image" | null | undefined;
} & import("class-variance-authority/types").ClassProp) | undefined) => string;
declare function AttachmentMedia({ className, variant, ...props }: React.ComponentProps<"div"> & VariantProps<typeof attachmentMediaVariants>): React.JSX.Element;
declare function AttachmentContent({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
declare function AttachmentTitle({ className, ...props }: React.ComponentProps<"span">): React.JSX.Element;
declare function AttachmentDescription({ className, ...props }: React.ComponentProps<"span">): React.JSX.Element;
declare function AttachmentActions({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
declare function AttachmentAction({ className, variant, size, ...props }: React.ComponentProps<typeof Button>): React.JSX.Element;
declare function AttachmentTrigger({ className, asChild, type, ...props }: React.ComponentProps<"button"> & {
    asChild?: boolean;
}): React.JSX.Element;
declare function AttachmentGroup({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
export { Attachment, AttachmentGroup, AttachmentMedia, AttachmentContent, AttachmentTitle, AttachmentDescription, AttachmentActions, AttachmentAction, AttachmentTrigger, };
