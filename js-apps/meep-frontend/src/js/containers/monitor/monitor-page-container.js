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
import _ from 'lodash';
import { Grid, GridCell, GridInner } from '@rmwc/grid';
import { Select } from '@rmwc/select';
import { Elevation } from '@rmwc/elevation';
import { Button } from '@rmwc/button';
import Iframe from 'react-iframe';
import HeadlineBar from '../../components/headline-bar';
import { ListEditPane } from './list-edit-pane';
import IDConfirmDialog from '../../components/dialogs/id-confirm-dialog';

import {
  uiChangeCurrentDialog
} from '../../state/ui';

import { deepCopy } from '../../util/object-util';

import { pipe, filter } from '../../util/functional';

import {
  changeDashboard,
  changeDashboardOptions,
  changeEditedDashboardOptions,
  resetDashboardOptions
} from '../../state/monitor';

import {
  MON_DASHBOARD_SELECT,
  MON_DASHBOARD_IFRAME,
  IDC_DIALOG_DELETE_DASHBOARD_LIST,
  IDC_DIALOG_RESET_DASHBOARD_LIST
} from '../../meep-constants';

const grafanaUrl = 'http://' + location.hostname + ':30009';

const DashboardContainer = props => {
  if (!props.dashboardUrl) {
    return null;
  }
  return (
    <Grid style={{ width: '100%', height: '100%' }}>
      <GridInner style={{ width: '100%', height: '100%' }}>
        <GridCell span={12} style={styles.inner}>
          <Elevation
            className="component-style"
            z={2}
            style={{
              width: '100%',
              height: '100%',
              display: 'flex',
              flexDirection: 'column'
            }}
          >
            <div style={{ flex: 1, padding: 10 }}>
              <div data-cy={MON_DASHBOARD_IFRAME} style={{ height: '100%' }}>
                <Iframe
                  url={props.dashboardUrl}
                  id="myId"
                  display="initial"
                  position="relative"
                  allowFullScreen
                  width='100%'
                  height='100%'
                />
              </div>
            </div>
          </Elevation>
        </GridCell>
      </GridInner>
    </Grid>
  );
};

const EditModeButton = ({ isEditMode, startEditMode }) => {
  return (
    <Button
      raised
      disabled={isEditMode()}
      style={styles.button}
      onClick={startEditMode}
    >
      EDIT
    </Button>
  );
};

const MonitorPageHeadlineBar = props => {

  const dashboardViewList = ((dashboardOptions) =>  {
    var dashboardList = [];
    if (dashboardOptions) {
      for (var i = 0; i < dashboardOptions.length; i++) {
        var dashboard = dashboardOptions[i];
        if (dashboard.label !== '') {
          dashboardList.push(dashboard.label);
        }
      }
    }
    return dashboardList;
  })(props.dashboardOptions);

  return (
    <div style={{ width: '100%' }}>
      <Grid style={styles.headlineGrid}>
        <GridCell span={12}>
          <Elevation className="component-style" z={2} style={styles.headline}>
            <GridInner>
              <GridCell align={'middle'} span={5}>
                <HeadlineBar
                  titleLabel="Deployed Scenario"
                  scenarioName={props.scenarioName}
                />
              </GridCell>
              <GridCell span={4} align={'middle'}>
                <Select
                  style={{ width: '100%' }}
                  label="Dashboard"
                  disabled={props.dashboardSelectDisabled}
                  outlined
                  options={dashboardViewList}
                  onChange={props.onChangeDashboard}
                  value={props.currentDashboard}
                  data-cy={MON_DASHBOARD_SELECT}
                />
              </GridCell>
              <GridCell span={3} align={'middle'}>
                <div align={'right'}>
                  <EditModeButton
                    isEditMode={props.isEditMode}
                    startEditMode={props.startEditMode}
                    cancelEditMode={props.cancelEditMode}
                  />
                  <Button
                    raised
                    style={styles.button}
                    onClick={() => window.open(grafanaUrl, '_blank')}
                  >
                  OPEN GRAFANA
                  </Button>
                </div>
              </GridCell>
            </GridInner>
          </Elevation>
        </GridCell>
      </Grid>
    </div>
  );
};

const MainPane = props => {
  if (props.editedDashboardOptions) {
    return (
      <Elevation z={4} style={{ marginBottom: 10, padding: 10 }}>
        <ListEditPane
          items={props.editedDashboardOptions}
          cancelEditMode={props.cancelEditMode}
          saveItems={props.saveDashboards}
          resetItems={props.resetDashboards}
          addItem={props.addOption}
          deleteItems={props.deleteOptions}
          itemLabelLabel={'Dashboard Name'}
          itemValueLabel={'Dashboard Url'}
          updateItemLabel={props.updateOptionLabel}
          updateItemValue={props.updateOptionValue}
          updateItemSelection={props.updateOptionSelection}
          canDelete={props.canDelete}
        />
      </Elevation>
    );
  } else {
    return <DashboardContainer dashboardUrl={props.currentDashboardUrl} />;
  }
};

class MonitorPageContainer extends Component {
  constructor(props) {
    super(props);
  }

  handleSelectionChange(e) {
    this.props.changeDashboard(e.target.value);
  }

  removeSelectedFlags() {
    const options = _.map(this.props.editedDashboardOptions, op => {
      return { label: op.label, value: op.value };
    });
    return options;
  }

  startEditMode() {
    const opts = JSON.parse(JSON.stringify(this.props.dashboardOptions));
    const options = _.map(opts, (op, index) => {
      return { ...op, data: { selected: false }, index: index };
    });
    this.props.changeEditedDashboardOptions(options);
  }

  cancelEditMode() {
    this.props.changeEditedDashboardOptions(null);
  }

  isEditMode() {
    return this.props.editedDashboardOptions !== null;
  }

  addOption() {
    const newOptions = [
      ...this.props.editedDashboardOptions,
      {
        label: '',
        value: '',
        data: { selected: false },
        index: this.props.editedDashboardOptions.length
      }
    ];
    this.props.changeEditedDashboardOptions(newOptions);
  }

  updateOptionAttribute(index, attribute, value) {
    let options = deepCopy(this.props.editedDashboardOptions);
    let option = { ...options[index], [attribute]: value };
    options[index] = option;
    this.props.changeEditedDashboardOptions(options);
  }

  performDeleteOptions() {
    const isNotSelected = option => !option.data.selected;

    const options = pipe(
      filter(isNotSelected),
      deepCopy
    )(this.props.editedDashboardOptions);

    this.props.changeEditedDashboardOptions(options);
  }

  performResetOptions() {
    this.props.resetDashboardOptions();
  }

  isOptionSelected(option) {
    return _.includes(this.state.selectedIndices, option.index);
  }

  canDelete() {
    if (!this.props.editedDashboardOptions) {
      return false;
    }

    let someSelected = _.reduce(
      this.props.editedDashboardOptions,
      (acc, option) => acc || option.data.selected,
      false
    );

    return someSelected;
  }

  showDialog(id) {
    this.props.showDialog(id);
  }

  closeDialog() {
    this.showDialog(null);
  }

  saveDashboards() {
    const options = deepCopy(this.props.editedDashboardOptions);
    this.props.changeDashboardOptions(options);
    this.props.changeEditedDashboardOptions(null);
  }



  render() {
    const showInExecStr = '<exec>';
    const passVarsStr = '<vars>';

    // Retrieve dashboard URL
    const currentDashboardUrl = (() => {
      var url = '';
      if (this.props.dashboardOptions) {
        for (var i=0; i < this.props.dashboardOptions.length; i++) {
          var dashboard = this.props.dashboardOptions[i];
          if (dashboard.label === this.props.currentDashboard) {
            url = dashboard.value;
            url = url.replace(showInExecStr, '');
            url = url.replace(passVarsStr, '');
            break;
          }
        }
      }
      return url;
    })();

    return (
      <div style={{ width: '100%', height: '100%' }}>
        <IDConfirmDialog
          title="Delete selected dashboards"
          open={this.props.currentDialog === IDC_DIALOG_DELETE_DASHBOARD_LIST}
          onClose={() => {
            this.closeDialog();
          }}
          onSubmit={() => this.performDeleteOptions()}
        />
        <IDConfirmDialog
          title="Reset default dashboards"
          open={this.props.currentDialog === IDC_DIALOG_RESET_DASHBOARD_LIST}
          onClose={() => {
            this.closeDialog();
          }}
          onSubmit={() => this.performResetOptions()}
        />
        <MonitorPageHeadlineBar
          scenarioName={this.props.scenarioName}
          onChangeDashboard={e => this.handleSelectionChange(e)}
          dashboardSelectDisabled={this.props.editedDashboardOptions !== null}
          dashboardOptions={this.props.dashboardOptions}
          currentDashboard={this.props.currentDashboard}
          isEditMode={() => this.isEditMode()}
          startEditMode={() => this.startEditMode()}
          open={this.props.mainDrawerOpen}
        />
        <MainPane
          editedDashboardOptions={this.props.editedDashboardOptions}
          currentDashboardUrl={currentDashboardUrl}
          cancelEditMode={() => this.cancelEditMode()}
          saveDashboards={() => this.saveDashboards()}
          resetDashboards={() => this.showDialog(IDC_DIALOG_RESET_DASHBOARD_LIST)}
          addOption={() => this.addOption()}
          deleteOptions={() => this.showDialog(IDC_DIALOG_DELETE_DASHBOARD_LIST)}
          updateOptionLabel={(index, value) =>
            this.updateOptionAttribute(index, 'label', value)
          }
          updateOptionValue={(index, value) =>
            this.updateOptionAttribute(index, 'value', value)
          }
          updateOptionSelection={(index, value) =>
            this.updateOptionAttribute(index, 'data', { selected: value })
          }
          canDelete={() => this.canDelete()}
        />
      </div>
    );
  }
}

const styles = {
  headlineGrid: {
    marginBottom: 10
  },
  headline: {
    paddingTop: 10,
    paddingRight: 10,
    paddingBottom: 10,
    paddingLeft: 10,
    marginBotton: 25
  },
  inner: {
    height: '100%'
  },
  page: {
    height: 1500,
    marginBottom: 10
  },
  cfgTable: {
    marginTop: 20,
    padding: 10
  },
  button: {
    color: 'white',
    marginRight: 5
  },
  editListItemCell: {
    padding: 5
  }
};

const mapStateToProps = state => {
  return {
    automaticRefresh: state.ui.automaticRefresh,
    devMode: state.ui.devMode,
    page: state.ui.page,
    scenarioName: state.exec.scenario.name,
    currentDashboard: state.monitor.currentDashboard,
    dashboardOptions: state.monitor.dashboardOptions,
    editedDashboardOptions: state.monitor.editedDashboardOptions,
    currentDialog: state.ui.currentDialog
  };
};

const mapDispatchToProps = dispatch => {
  return {
    changeDashboard: label => dispatch(changeDashboard(label)),
    changeEditedDashboardOptions: mode => dispatch(changeEditedDashboardOptions(mode)),
    changeDashboardOptions: mode => dispatch(changeDashboardOptions(mode)),
    resetDashboardOptions: () => dispatch(resetDashboardOptions()),
    showDialog: type => dispatch(uiChangeCurrentDialog(type))
  };
};

const ConnectedMonitorPageContainer = connect(
  mapStateToProps,
  mapDispatchToProps
)(MonitorPageContainer);

export default ConnectedMonitorPageContainer;
