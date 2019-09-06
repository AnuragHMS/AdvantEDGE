/**
 * MEEP Controller REST API
 * Copyright (c) 2019 InterDigital Communications, Inc. All rights reserved. The information provided herein is the proprietary and confidential information of InterDigital Communications, Inc. 
 *
 * OpenAPI spec version: 1.0.0
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 *
 * Swagger Codegen version: 2.3.1
 *
 * Do not edit the class manually.
 *
 */

(function(root, factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as an anonymous module.
    define(['ApiClient', 'model/ServiceMap'], factory);
  } else if (typeof module === 'object' && module.exports) {
    // CommonJS-like environments that support module.exports, like Node.
    module.exports = factory(require('../ApiClient'), require('./ServiceMap'));
  } else {
    // Browser globals (root is window)
    if (!root.MeepControllerRestApi) {
      root.MeepControllerRestApi = {};
    }
    root.MeepControllerRestApi.ClientServiceMap = factory(root.MeepControllerRestApi.ApiClient, root.MeepControllerRestApi.ServiceMap);
  }
}(this, function(ApiClient, ServiceMap) {
  'use strict';




  /**
   * The ClientServiceMap model module.
   * @module model/ClientServiceMap
   * @version 1.0.0
   */

  /**
   * Constructs a new <code>ClientServiceMap</code>.
   * Client-specific list of mappings of exposed port to internal service
   * @alias module:model/ClientServiceMap
   * @class
   */
  var exports = function() {
    var _this = this;



  };

  /**
   * Constructs a <code>ClientServiceMap</code> from a plain JavaScript object, optionally creating a new instance.
   * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
   * @param {Object} data The plain JavaScript object bearing properties of interest.
   * @param {module:model/ClientServiceMap} obj Optional instance to populate.
   * @return {module:model/ClientServiceMap} The populated <code>ClientServiceMap</code> instance.
   */
  exports.constructFromObject = function(data, obj) {
    if (data) {
      obj = obj || new exports();

      if (data.hasOwnProperty('client')) {
        obj['client'] = ApiClient.convertToType(data['client'], 'String');
      }
      if (data.hasOwnProperty('serviceMap')) {
        obj['serviceMap'] = ApiClient.convertToType(data['serviceMap'], [ServiceMap]);
      }
    }
    return obj;
  }

  /**
   * Unique external client identifier
   * @member {String} client
   */
  exports.prototype['client'] = undefined;
  /**
   * @member {Array.<module:model/ServiceMap>} serviceMap
   */
  exports.prototype['serviceMap'] = undefined;



  return exports;
}));

