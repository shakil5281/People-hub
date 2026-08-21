import * as React from "react";
declare function MessageGroup({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
declare function Message({ className, align, ...props }: React.ComponentProps<"div"> & {
    align?: "start" | "end";
}): React.JSX.Element;
declare function MessageAvatar({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
declare function MessageContent({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
declare function MessageHeader({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
declare function MessageFooter({ className, ...props }: React.ComponentProps<"div">): React.JSX.Element;
export { MessageGroup, Message, MessageAvatar, MessageContent, MessageFooter, MessageHeader, };
