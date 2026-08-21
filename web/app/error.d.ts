import * as React from "react";
export default function Error({ error, reset, }: {
    error: Error & {
        digest?: string;
    };
    reset: () => void;
}): React.JSX.Element;
