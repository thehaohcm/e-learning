// Requests from the UI to the app task (network work never runs in the UI task).
#pragma once

#include <stdbool.h>

typedef enum {
    APP_REQ_SYNC,        // send answers + download the next batch
    APP_REQ_PUSH,        // send answers only (end of a study round)
    APP_REQ_WIFI_SETUP,  // forget Wi-Fi, restart into the setup access point
    APP_REQ_UNPAIR,      // forget the device token (link again with a new code)
    APP_REQ_RETRY,       // retry connecting
} app_req_t;

void app_request(app_req_t r);

/** Status shown in the header / home screen. */
typedef struct {
    bool online;
    bool syncing;
    int pending;        // answers not sent yet
    char user[48];
    char message[64];   // last sync problem, "" if fine
} app_status_t;

const app_status_t *app_status(void);
