//
//  ViewController.m
//  ANRDemo
//
//  Created by Brent Shu on 2018/8/14.
//  Copyright © 2018年 sky. All rights reserved.
//

#import "ViewController.h"

@interface ViewController ()

@end

@implementation ViewController

- (void)viewDidLoad {
    [super viewDidLoad];
    // Do any additional setup after loading the view, typically from a nib.
    [self test];
}

- (void)test {
    double now = CACurrentMediaTime();
    while ((CACurrentMediaTime() - now) * 1000 < 101) {
    }
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(1 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
        [self test];
    });
}


- (void)didReceiveMemoryWarning {
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}


@end
