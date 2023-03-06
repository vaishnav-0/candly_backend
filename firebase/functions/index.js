const functions = require("firebase-functions");

// // Create and Deploy Your First Cloud Functions
// // https://firebase.google.com/docs/functions/write-firebase-functions
//
// exports.helloWorld = functions.https.onRequest((request, response) => {
//   functions.logger.info("Hello logs!", {structuredData: true});
//   response.send("Hello from Firebase!");
// });

const functions = require("firebase-functions");
const admin = require('firebase-admin');
require("firebase-functions/lib/logger/compat");


admin.initializeApp();
const auth = admin.auth();


exports.createUser = functions.auth.user().onCreate(async (user) => {
    const userData = {
        name: user.displayName,
        email: user.email,
        phoneNo: user.phoneNumber,
        photoURL: user.photoURL,
        uid: user.uid,
    };

    

    //
});
