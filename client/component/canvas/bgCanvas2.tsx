import React from 'react'
import * as BABYLON from 'babylonjs';

const width = 1344
const height = 600
const rectX = width/2
const rectY = height/2


export default class bgCanvas2 extends React.Component{

    bgCanvas:React.RefObject<HTMLCanvasElement>
    bgCtx:CanvasRenderingContext2D

    constructor(props:any){
        super(props)
        this.bgCanvas = React.createRef()
        this.bgCtx = null
    }

    componentDidMount(){
        var engine = new BABYLON.Engine(this.bgCanvas.current, true, {preserveDrawingBuffer: true, stencil: true});
        // CreateScene function that creates and return the scene
        var createScene = function(){
            // Create a basic BJS Scene object
            var scene = new BABYLON.Scene(engine);
            // Create a FreeCamera, and set its position to {x: 0, y: 5, z: -10}
            var camera = new BABYLON.FreeCamera('camera1', new BABYLON.Vector3(0, 5, -10), scene);
            // Target the camera to scene origin
            camera.setTarget(BABYLON.Vector3.Zero());
            // Attach the camera to the canvas
            camera.attachControl(this.bgCanvas.current, false);
            // Create a basic light, aiming 0, 1, 0 - meaning, to the sky
            var light = new BABYLON.HemisphericLight('light1', new BABYLON.Vector3(0, 1, 0), scene);
            // Create a built-in "sphere" shape; its constructor takes 6 params: name, segment, diameter, scene, updatable, sideOrientation
            var sphere = BABYLON.Mesh.CreateSphere('sphere1', 16, 2, scene, false, BABYLON.Mesh.FRONTSIDE);
            // Move the sphere upward 1/2 of its height
            sphere.position.y = 1;
            // Create a built-in "ground" shape; its constructor takes 6 params : name, width, height, subdivision, scene, updatable
            var ground = BABYLON.Mesh.CreateGround('ground1', 6, 6, 2, scene, false);
            // Return the created scene
            return scene;
        }.bind(this)
        // call the createScene function
        var scene = createScene();
        // run the render loop
        engine.runRenderLoop(function(){
            scene.render();
        });
    }

    
    draw = (alphaPath:Path2D) =>{

    }

    render(){
        return (
            <div>
                <canvas width={width} height={height} className="bgCanvas" ref={this.bgCanvas} />
            </div>
        )
    }
}