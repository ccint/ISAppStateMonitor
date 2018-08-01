//
//  ViewController.m
//  ANRTest
//
//  Created by Brent Shu on 2018/5/10.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import "ViewController.h"

@interface ViewController ()

@end

@implementation ViewController

- (void)viewDidLoad {
    [super viewDidLoad];
    // Do any additional setup after loading the view, typically from a nib.
    [self testInAnotherOne];
}

- (void)testInAnotherOne {
    double now = CACurrentMediaTime();
    while ((CACurrentMediaTime() - now) * 1000 < 1002) {
    }
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.5 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
        [self testInAnotherOne];
    });
}


- (void)didReceiveMemoryWarning {
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}


@end
