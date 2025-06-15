// $app.rootCmd.addCommand(new Command({
//     use: "hello",
//     run: (cmd, args) => {
//         console.log("Hello world!")
//     },
// }))

onBootstrap((e) => {
    e.next()

    console.log("Hello !") // <-- name will be undefined inside the handler
})