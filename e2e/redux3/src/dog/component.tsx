import * as React from "react";
import { connect } from "react-redux";
import { Dispatch, bindActionCreators  } from 'redux';

import * as DogActions from './actions';
import { RootState } from '../rootState';


// TODO move the logo
let logo = require('../logo.svg')


interface PropsFromState {
  fetching: boolean;
  imgSrc: string;
  error: string;
}

interface PropsFromDispatch {
  onRequestDog: () => void;
  cancelRequestDog: () => void;
  // another example:
  // onRequestDog: (value:string) => void;
}

interface PropsFromComponent {
  parentPropsExample: string
}

interface ReduxProps extends PropsFromState, PropsFromDispatch {}
interface Props extends ReduxProps, PropsFromComponent {}


interface ComponentLocalState {
  readonly localStateExample: string,
}

class Dog extends React.Component<Props> {
  state: ComponentLocalState;
  constructor(props: Props) {
    super(props);

    this.state = {
      localStateExample: "money",
    };

    this.handleClick = this.handleClick.bind(this)
  }

  componentWillReceiveProps() {
  }

  handleClick(): void {
    this.state.localStateExample == 'money' ?
      this.setState({localStateExample: 'fame'}) :
      this.setState({localStateExample: 'money'});
  }

  render() {
    const { fetching, imgSrc, error, onRequestDog, cancelRequestDog } = this.props;
    return (
      <div className="App">
        <header className="App-header">
          <img src={imgSrc || logo} className="App-logo" alt="logo" />
          <h1 className="App-title">Welcome to Dog Saga</h1>
        </header>
        <div style={{"marginTop":"-20px", paddingTop: '30px', "background": "#f1f1f1", "paddingBottom":"50px"}}>
          {imgSrc ? (
            <p className="App-intro">Keep clicking for new dogs</p>
          ) : (
            <p className="App-intro">Replace the React icon with a dog!</p>
          )}

          {fetching ? (
            <button disabled>Fetching...</button>
          ) : (
            <button onClick={onRequestDog}>Request a Dog</button>
          )}

          {error && <p style={{ color: "red" }}>Uh oh - something went wrong!</p>}

          <button onClick={cancelRequestDog}>Cancel API call</button>
          <br />
          <button onClick={this.handleClick}>{this.state.localStateExample}</button>
      </div>
      </div>
    ) as React.ReactNode;
  }
}


function mapStateToProps(state: RootState, ownProps: PropsFromComponent): PropsFromState {
  console.log('ownProps: ', ownProps);
  return {
    fetching: state.dog.fetching,
    imgSrc: state.dog.imgSrc,
    error: state.dog.error
  }
};

function mapDispatchToProps(dispatch: Dispatch<RootState>): PropsFromDispatch {
  return bindActionCreators({
    onRequestDog: DogActions.dogRequest,
    cancelRequestDog: DogActions.dogCancel,
    }, dispatch);
}

export default connect<PropsFromState, PropsFromDispatch, PropsFromComponent>(mapStateToProps, mapDispatchToProps)(Dog);
