import * as React from "react";

import { Hello } from "./Hello";
import { Login } from "./Login";

// 'HelloProps' describes the shape of props.
// State is never set so we use the '{}' type.
export class Root extends React.Component<{}, {}> {
    render() {
        return <div>
            <h1>Community Chess</h1>
            <Hello compiler="TypeScript" framework="React" ></Hello>
            <Login></Login>
        </div>;
    }
}