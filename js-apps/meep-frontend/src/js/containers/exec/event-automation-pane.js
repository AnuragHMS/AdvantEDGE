/*
 * Copyright (c) 2019  InterDigital Communications, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { connect } from 'react-redux';
import React, { Component } from 'react';
import { Button } from '@rmwc/button';
import { Checkbox } from '@rmwc/checkbox';
import { Grid, GridInner, GridCell } from '@rmwc/grid';
import { Typography } from '@rmwc/typography';

import {
  uiExecChangeAutomationMovementMode,
  uiExecChangeAutomationMobilityMode,
  uiExecChangeAutomationPoasInRangeMode
} from '../../state/ui';

const AUTO_TYPE_MOVEMENT = 'MOVEMENT';
const AUTO_TYPE_MOBILITY = 'MOBILITY';
const AUTO_TYPE_POAS_IN_RANGE = 'POAS-IN-RANGE';

class EventAutomationPane extends Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  componentDidMount() {
    this.refreshAutomationModes();
  }

  componentDidUpdate(prevProps) {
    if (this.props.sandbox !== prevProps.sandbox) {
      this.refreshAutomationModes();
    }
  }

  // Sync automation states with backend
  refreshAutomationModes() {
    this.props.api.getAutomationState((error, data) => {
      if (!error) {
        for (var i = 0; i < data.states.length; i++) {
          let mode = data.states[i].active ? true : false;
          switch (data.states[i].type) {
          case AUTO_TYPE_MOVEMENT:
            this.props.changeAutomationMovementMode(mode);
            break;
          case AUTO_TYPE_MOBILITY:
            this.props.changeAutomationMobilityMode(mode);
            break;
          case AUTO_TYPE_POAS_IN_RANGE:
            this.props.changeAutomationPoasInRangeMode(mode);
            break;
          default:
            break;
          }
        }
      }
    });
  }

  setMovementMode(mode) {
    this.props.changeAutomationMovementMode(mode);
    this.props.api.setAutomationStateByName(AUTO_TYPE_MOVEMENT, mode, (error) => {
      if (error) {
        // TODO consider showing an alert
        // console.log(error);
      }
    });
  }

  setMobilityMode(mode) {
    this.props.changeAutomationMobilityMode(mode);
    this.props.api.setAutomationStateByName(AUTO_TYPE_MOBILITY, mode, (error) => {
      if (error) {
        // TODO consider showing an alert
        // console.log(error);
      }
    });
  }

  setPoasInRangeMode(mode) {
    this.props.changeAutomationPoasInRangeMode(mode);
    this.props.api.setAutomationStateByName(AUTO_TYPE_POAS_IN_RANGE, mode, (error) => {
      if (error) {
        // TODO consider showing an alert
        // console.log(error);
      }
    });
  }

  render() {
    if (this.props.hide) {
      return null;
    }

    return (
      <div style={{ padding: 10 }}>
        <div style={styles.block}>
          <Typography use="headline6">Event Automation</Typography>
        </div>

        <Grid style={{ marginBottom: 20 }}>
          <GridCell span={12}>
            <Checkbox
              checked={this.props.automationMovementMode}
              onChange={e => this.setMovementMode(e.target.checked)}
            >
              Movement
            </Checkbox>
          </GridCell>
          <GridCell span={12}>
            <Checkbox
              checked={this.props.automationMobilityMode}
              onChange={e => this.setMobilityMode(e.target.checked)}
            >
              Mobility
            </Checkbox>
          </GridCell>
          <GridCell span={12}>
            <Checkbox
              checked={this.props.automationPoasInRangeMode}
              onChange={e => this.setPoasInRangeMode(e.target.checked)}
            >
              POAs in range
            </Checkbox>
          </GridCell>
        </Grid>

        <Grid style={{ marginTop: 10 }}>
          <GridInner>
            <GridCell span={12}>
              <Button
                outlined
                style={styles.button}
                onClick={this.props.onClose}
              >
                Close
              </Button>
            </GridCell>
          </GridInner>
        </Grid>
      </div>
    );
  }
}

const styles = {
  button: {
    marginRight: 10
  },
  block: {
    marginBottom: 20
  },
  field: {
    marginBottom: 10
  }
};

const mapStateToProps = state => {
  return {
    automationMovementMode: state.ui.automationMovementMode,
    automationMobilityMode: state.ui.automationMobilityMode,
    automationPoasInRangeMode: state.ui.automationPoasInRangeMode,
    sandbox: state.ui.sandbox
  };
};

const mapDispatchToProps = dispatch => {
  return {
    changeAutomationMovementMode: mode => dispatch(uiExecChangeAutomationMovementMode(mode)),
    changeAutomationMobilityMode: mode => dispatch(uiExecChangeAutomationMobilityMode(mode)),
    changeAutomationPoasInRangeMode: mode => dispatch(uiExecChangeAutomationPoasInRangeMode(mode))
  };
};

const ConnectedEventAutomationPane = connect(
  mapStateToProps,
  mapDispatchToProps
)(EventAutomationPane);

export default ConnectedEventAutomationPane;
