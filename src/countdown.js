import React from "react";

class Countdown extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            diffSeconds: this.secondsRemaining()
        };
        setInterval(() => {this.setState({diffSeconds:this.secondsRemaining()})}, 1000);
    }

    secondsRemaining() {
        return Math.floor(Math.max((this.props.endTimeMs - (new Date()).getTime()) / 1000, 0));
    }

    render() {
        return (
            <div id="NextMoveTimer"><span id="Number">{this.state.diffSeconds}</span></div>
        );
    }
}

export default Countdown;